package operations

import (
	"context"
	"fmt"
	"os"

	"github.com/go-viper/mapstructure/v2"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type firewallCmdParameters struct {
	Arguments              []string `mapstructure:"arguments"`
	LogCommandInStdStreams bool     `mapstructure:"log_command_in_std_streams"`
}

type FirewallCmdOperation struct {
	rawParameters map[string]any
	parameters    firewallCmdParameters
}

// Name implements [Operation].
func (s *FirewallCmdOperation) Name() string {
	return string(OperationNameFirewallCmd)
}

// PreCheck implements [Operation].
func (s *FirewallCmdOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// Run implements [Operation].
func (s *FirewallCmdOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	cmdInput := shellexecutors.CommandInput{
		Mode:   shellexecutors.ModeSync,
		Stdout: env.Journaler,
		Stderr: env.Journaler,
		Command: shellexecutors.Command{
			Executable: "firewall-cmd",
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

// SetParameters implements [Operation].
func (s *FirewallCmdOperation) SetParameters(parameters map[string]any) error {
	s.rawParameters = parameters
	if err := mapstructure.Decode(s.rawParameters, &s.parameters); err != nil {
		return fmt.Errorf("failed to parse parameters for 'firewall-cmd': %w", err)
	}

	return nil
}

func NewFirewallCmdOperation() *FirewallCmdOperation {
	return &FirewallCmdOperation{}
}
