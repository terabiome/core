package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/terabiome/core/internal/registry"
	"github.com/terabiome/core/internal/runner"
	manifestcontracts "github.com/terabiome/core/pkg/contracts/manifest"
	"github.com/terabiome/core/pkg/logging"
	"go.yaml.in/yaml/v4"
)

const shellSubcommand subcommand = "shell"

type shellCLIInput struct {
	YAMLConfigPath string `validate:"required"`
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

	var uow manifestcontracts.UnitOfWork
	if err = yaml.Unmarshal(content, &uow); err != nil {
		return fmt.Errorf("%s: %w", errMsgFailedToDecodeYAMLConfig, err)
	}

	shellRunner := runner.NewShellRunner(
		logging.GetGlobalSlogger(),
		registry.NewDefault(),
	)

	if shellRunner == nil {
		return errors.New(errMsgFailedToInitializeShellRunner)
	}

	return shellRunner.Run(ctx, uow)
}
