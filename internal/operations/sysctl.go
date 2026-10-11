package operations

import (
	"context"
	"fmt"
	"os"

	"github.com/go-viper/mapstructure/v2"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type sysctlCmdParameters struct {
	Arguments              []string `mapstructure:"arguments"`
	LogCommandInStdStreams bool     `mapstructure:"log_command_in_std_streams"`
}

type SysctlCmdOperation struct {
	rawParameters map[string]any
	parameters    sysctlCmdParameters
}

func (s *SysctlCmdOperation) Name() string {
	return string(OperationNameSysctl)
}

// PreCheck: Aggressive Gatekeeping
func (s *SysctlCmdOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

func (s *SysctlCmdOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	cmdInput := shellexecutors.CommandInput{
		Mode:   shellexecutors.ModeSync,
		Stdout: env.Journaler,
		Stderr: env.Journaler,
		Command: shellexecutors.Command{
			Executable: "sysctl",
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

func (s *SysctlCmdOperation) SetParameters(parameters map[string]any) error {
	s.rawParameters = parameters
	if err := mapstructure.Decode(s.rawParameters, &s.parameters); err != nil {
		return fmt.Errorf("failed to parse parameters for 'sysctl': %w", err)
	}
	return nil
}

func NewSysctlCmdOperation() *SysctlCmdOperation {
	return &SysctlCmdOperation{}
}
