package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	apihttp "github.com/terabiome/core/internal/api"
	"github.com/terabiome/core/internal/config"
	yamlcontracts "github.com/terabiome/core/internal/contracts/yaml"
	"github.com/terabiome/core/internal/database"
	httphandlers "github.com/terabiome/core/internal/handlers/http"
	"github.com/terabiome/core/internal/runner"
	"github.com/terabiome/core/internal/singletons"
	"github.com/terabiome/core/pkg/logging"
	"go.yaml.in/yaml/v4"
)

const (
	errMsgFailedToReadYAMLConfig        = "failed to read YAML config"
	errMsgFailedToDecodeYAMLConfig      = "failed to decode YAML config"
	errMsgFailedToInitializeShellRunner = "failed to initialize shell runner"
)

type subcommand string

const (
	shellSubcommand    subcommand = "shell"
	apiSubcommand      subcommand = "api"
	consumerSubcommand subcommand = "consumer"
)

type shellCLIInput struct {
	YAMLConfigPath string `validate:"required"`
}

type apiCLIInput struct {
	Addr string
}

type cliInput struct {
	subcommand subcommand
	shell      shellCLIInput
	api        apiCLIInput
}

func NewCLIInput() *cliInput {
	return &cliInput{
		shell: shellCLIInput{},
	}
}

var mainValidator = validator.New()

func (i *cliInput) Parse() error {
	shellCmd := flag.NewFlagSet(string(shellSubcommand), flag.ExitOnError)
	shellCmd.StringVar(&i.shell.YAMLConfigPath, "yaml-config-path", "", "Path to YAML config file")

	apiCmd := flag.NewFlagSet(string(apiSubcommand), flag.ExitOnError)
	apiCmd.StringVar(&i.api.Addr, "addr", ":8080", "Address for the HTTP API")

	parseAndValidateFn := func(fs *flag.FlagSet, sPtr any) error {
		return errors.Join(
			fs.Parse(os.Args[2:]),
			mainValidator.Struct(sPtr),
		)
	}

	// atp, it is guaranteed to have > 1 argument
	switch subcommand(os.Args[1]) {
	case shellSubcommand:
		i.subcommand = shellSubcommand
		if err := parseAndValidateFn(shellCmd, &i.shell); err != nil {
			return err
		}
	case apiSubcommand:
		i.subcommand = apiSubcommand
		if err := parseAndValidateFn(apiCmd, &i.api); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported subcommand: %s", os.Args[1])
	}

	return nil
}

func processShellSubcommand(ctx context.Context, shellInput shellCLIInput) error {
	content, err := os.ReadFile(shellInput.YAMLConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s: file not found", errMsgFailedToReadYAMLConfig)
		}
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("%s: unprivileged", errMsgFailedToReadYAMLConfig)
		}
		return fmt.Errorf("%s: %w", errMsgFailedToReadYAMLConfig, err)
	}

	var uow yamlcontracts.UnitOfWork
	if err = yaml.Unmarshal(content, &uow); err != nil {
		return fmt.Errorf("%s: %w", errMsgFailedToDecodeYAMLConfig, err)
	}

	shellRunner := runner.NewShellRunner(
		logging.GetGlobalSlogger(),
	)

	if shellRunner == nil {
		return errors.New(errMsgFailedToInitializeShellRunner)
	}

	return shellRunner.Run(ctx, uow)
}

const (
	dispatchSweepLimit = 32
	dispatchSweepEvery = 5 * time.Second
)

func processAPISubcommand(ctx context.Context, apiInput apiCLIInput) error {
	if err := singletons.Init(); err != nil {
		return fmt.Errorf("failed to initialize singletons: %w", err)
	}
	defer database.Close()

	if err := singletons.InitDispatch(); err != nil {
		return fmt.Errorf("failed to initialize dispatch: %w", err)
	}
	defer func() {
		if err := singletons.CloseDispatch(); err != nil {
			logging.GetGlobalSlogger().Error("failed to close NATS", "error", err)
		}
	}()

	sweepCtx, stopSweep := context.WithCancel(ctx)
	defer stopSweep()
	sweepDone := make(chan struct{})
	go func() {
		defer close(sweepDone)
		singletons.Get().DispatchService.RunSweep(
			sweepCtx,
			dispatchSweepLimit,
			dispatchSweepEvery,
			config.Get().Dispatch.SweepPassTimeout,
		)
	}()

	server := &http.Server{
		Addr:              apiInput.Addr,
		Handler:           apihttp.SetupMux(httphandlers.NewUnitOfWork(singletons.Get().UnitOfWorkService)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logging.GetGlobalSlogger().Info("api listening", "addr", apiInput.Addr)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-errCh:
	}

	// Stop after the current pass. Do not drain HTTP first, and do not block ListenAndServe on this loop.
	stopSweep()
	select {
	case <-sweepDone:
	case <-time.After(30 * time.Second):
		logging.GetGlobalSlogger().Error("dispatch sweep did not stop before timeout")
	}

	if serveErr != nil {
		return serveErr
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("expect subcommand, only 1 argument")
	}

	cliInput := NewCLIInput()
	if err := cliInput.Parse(); err != nil {
		log.Fatalf("failed to parse argument flags: %v\n", err)
	}

	singletons.InitLogger()

	// Ctrl + C or something to stop the command halfway
	mainCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var err error
	switch cliInput.subcommand {
	case shellSubcommand:
		if os.Getuid() != 0 {
			log.Fatalln("shell subcommand must be run as root/sudo")
		}
		err = processShellSubcommand(mainCtx, cliInput.shell)
	case apiSubcommand:
		err = processAPISubcommand(mainCtx, cliInput.api)
	}

	if err != nil {
		logging.GetGlobalSlogger().Error("execution failed",
			"subcommand", cliInput.subcommand,
			"error", err,
		)
	}

	log.Println("completed")
}
