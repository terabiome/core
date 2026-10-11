package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
	"github.com/terabiome/terabiome/internal/runner"
	"github.com/terabiome/terabiome/internal/singletons"
	"github.com/terabiome/terabiome/pkg/logging"
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

type cliInput struct {
	subcommand subcommand
	shell      shellCLIInput
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
	default:
		return fmt.Errorf("unsupported subcommand: %s", os.Args[1])
	}

	return nil
}

func processShellSubcommand(ctx context.Context, shellInput shellCLIInput) error {
	content, err := os.ReadFile(shellInput.YAMLConfigPath)
	if err != nil {
		if castedErr, ok := err.(*os.PathError); ok {
			switch castedErr {
			case fs.ErrNotExist:
				return fmt.Errorf("%s: file not found", errMsgFailedToReadYAMLConfig)
			case fs.ErrPermission:
				return fmt.Errorf("%s: unprivileged", errMsgFailedToReadYAMLConfig)
			}
		}
		return fmt.Errorf("%s: %w", errMsgFailedToReadYAMLConfig, err)
	}

	var uow yamlcontracts.UnitOfWork
	if err = yaml.Unmarshal(content, &uow); err != nil {
		return fmt.Errorf("%s: %w", errMsgFailedToDecodeYAMLConfig, err)
	}

	shellRunner := runner.NewShellRunner(
		logging.GetGlobalSlogger(),
		singletons.GetWorkflowService(),
	)

	if shellRunner == nil {
		return errors.New(errMsgFailedToInitializeShellRunner)
	}

	return shellRunner.Run(ctx, uow)
}

func main() {
	if os.Getuid() != 0 {
		log.Fatalln("this script must be run as root/sudo")
	}

	if len(os.Args) < 2 {
		log.Fatalln("expect subcommand, only 1 argument")
	}

	cliInput := NewCLIInput()
	if err := cliInput.Parse(); err != nil {
		log.Fatalf("failed to parse argument flags: %v\n", err)
	}

	// injectable dependencies live here
	singletons.Init()

	// Ctrl + C or something to stop the command halfway
	mainCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var err error
	switch cliInput.subcommand {
	case shellSubcommand:
		err = processShellSubcommand(mainCtx, cliInput.shell)
	}

	if err != nil {
		logging.GetGlobalSlogger().Error("execution failed",
			"subcommand", cliInput.subcommand,
			"error", err,
		)
	}

	log.Println("completed")
}
