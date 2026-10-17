package operations

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type baseShellParameters struct {
	Executable             string   `mapstructure:"-"` // Hidden from mapstructure
	Arguments              []string `mapstructure:"arguments"`
	ArgumentsAsString      string   `mapstructure:"arguments_as_string"`
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

	if len(s.parameters.Arguments) > 1 && s.parameters.ArgumentsAsString != "" {
		return fmt.Errorf("conflicted arguments: can only be arguments or arguments_as_string present")
	}

	// argument array is guaranteed to be empty here
	if s.parameters.ArgumentsAsString != "" {
		s.parameters.Arguments = strings.Fields(s.parameters.ArgumentsAsString)
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
