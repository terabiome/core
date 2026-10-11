package operations

import (
	"context"
	"fmt"
	"os"

	"github.com/go-viper/mapstructure/v2"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type systemctlParameters struct {
	Arguments              []string `mapstructure:"arguments"`
	LogCommandInStdStreams bool     `mapstructure:"log_command_in_std_streams"`
}

type SystemctlOperation struct {
	parameters map[string]any
}

// Name implements [Operation].
func (s SystemctlOperation) Name() string {
	return "systemctl"
}

// PreCheck implements [Operation].
func (s SystemctlOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// Run implements [Operation].
func (s SystemctlOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	var parameters systemctlParameters
	if err := mapstructure.Decode(s.parameters, &parameters); err != nil {
		return fmt.Errorf("failed to parse parameters for 'systemctl': %w", err)
	}

	cmdInput := shellexecutors.CommandInput{
		Mode:   shellexecutors.ModeSync,
		Stdout: env.Journaler,
		Stderr: env.Journaler,
		Command: shellexecutors.Command{
			Executable: "systemctl",
			Arguments:  parameters.Arguments,
		},
	}
	if parameters.LogCommandInStdStreams {
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

// SetParameters implements [Operation].
func (s *SystemctlOperation) SetParameters(parameters map[string]any) error {
	s.parameters = parameters
	return nil
}

func NewSystemctlOperation() *SystemctlOperation {
	return &SystemctlOperation{}
}
