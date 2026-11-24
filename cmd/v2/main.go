package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/terabiome/core/internal/singletons"
	"github.com/terabiome/core/pkg/logging"
)

const (
	errMsgFailedToReadYAMLConfig        = "failed to read YAML config"
	errMsgFailedToDecodeYAMLConfig      = "failed to decode YAML config"
	errMsgFailedToInitializeShellRunner = "failed to initialize shell runner"
)

type subcommand string

const (
	consumerSubcommand subcommand = "consumer"
)

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
