package operations

import (
	"context"
	"fmt"
	"os"

	"github.com/go-viper/mapstructure/v2"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type baseShellParameters struct {
	Executable             string   `mapstructure:"-"`
	Arguments              []string `mapstructure:"arguments"`
	LogCommandInStdStreams bool     `mapstructure:"log_command_in_std_streams"`
}

type baseShellOperationOperation struct {
	rawParameters map[string]any
	parameters    baseShellParameters
}

func (s *baseShellOperationOperation) SetParameters(parameters map[string]any) error {
	s.rawParameters = parameters
	if err := mapstructure.Decode(s.rawParameters, &s.parameters); err != nil {
		return fmt.Errorf("failed to parse parameters for '%s': %w", s.parameters.Executable, err)
	}

	return nil
}

func (s *baseShellOperationOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	cmdInput := shellexecutors.CommandInput{
		Mode:   shellexecutors.ModeSync,
		Stdout: env.Journaler,
		Stderr: env.Journaler,
		Command: shellexecutors.Command{
			Executable: s.parameters.Executable,
			Arguments:  s.parameters.Arguments,
		},
	}
	if s.parameters.LogCommandInStdStreams {
		cmdInput.Stdout = os.Stdout
		cmdInput.Stderr = os.Stderr
	}

	output := env.Executor.Execute(ctx, &cmdInput)
	output.Wait()
	if output.Error != nil {
		return fmt.Errorf("execution error: %w", output.Error)
	}

	return nil
}
