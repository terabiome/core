package operations

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

var whitelistedCommands = map[string]*regexp.Regexp{
	"systemctl":    nil,
	"sysctl":       nil,
	"modprobe":     nil,
	"firewall-cmd": nil,
}

type shellParameters struct {
	Executable             string   `mapstructure:"-"`
	Arguments              []string `mapstructure:"-"`
	Command                string   `mapstructure:"command"`
	CommandSegments        []string `mapstructure:"command_segments"`
	LogCommandInStdStreams bool     `mapstructure:"log_command_in_std_streams"`
}

type ShellCmdOperation struct {
	rawParameters map[string]any
	parameters    shellParameters
}

// Name implements [Operation].
func (s *ShellCmdOperation) Name() string {
	return "shell-cmd"
}

// PreCheck implements [Operation].
func (s *ShellCmdOperation) PreCheck(env *OperationEnvironment) error {
	regex, exist := whitelistedCommands[s.parameters.Executable]
	if !exist {
		return fmt.Errorf("command is not allowed to run")
	}

	if regex != nil && !regex.MatchString(s.parameters.Command) {
		return fmt.Errorf("command does not match the format")
	}

	return nil
}

// SetParameters implements [Operation].
func (s *ShellCmdOperation) SetParameters(parameters map[string]any) error {
	s.rawParameters = parameters
	if err := mapstructure.Decode(s.rawParameters, &s.parameters); err != nil {
		return fmt.Errorf("failed to parse parameters for '%s': %w", s.parameters.Executable, err)
	}

	if len(s.parameters.CommandSegments) > 1 && s.parameters.Command != "" {
		return fmt.Errorf("conflicted arguments: can only be arguments or arguments_as_string present")
	}

	if s.parameters.Command != "" {
		s.parameters.CommandSegments = strings.Fields(s.parameters.Command)
	} else {
		s.parameters.Command = strings.Join(s.parameters.CommandSegments, " ")
	}

	if len(s.parameters.CommandSegments) < 1 {
		return fmt.Errorf("not enough segments in shell command")
	}

	s.parameters.Executable = s.parameters.CommandSegments[0]
	if len(s.parameters.CommandSegments) > 1 {
		s.parameters.Arguments = s.parameters.CommandSegments[1:]
	}

	return nil
}

// Run implements [Operation].
func (s *ShellCmdOperation) Run(ctx context.Context, env *OperationEnvironment) error {
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

func NewShellCmdOperation() *ShellCmdOperation {
	return &ShellCmdOperation{}
}
