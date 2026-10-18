package shell

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/terabiome/internal/operations"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

var whitelistedCommands = map[string][]regexp.Regexp{
	"systemctl":    {},
	"sysctl":       {},
	"modprobe":     {},
	"firewall-cmd": {},
}

type CmdOperation struct {
	rawParameters map[string]any
	parameters    cmdParameters
}

// Name implements [Operation].
func (op *CmdOperation) Name() string {
	return string(operations.OperationNameShellCmd)
}

// PreCheck implements [Operation].
func (op *CmdOperation) PreCheck(env *operations.OperationEnvironment) error {
	regexes, exist := whitelistedCommands[op.parameters.Executable]
	if !exist {
		return fmt.Errorf("command is not allowed to run")
	}

	if len(regexes) > 0 {
		// greedy, match 1st and return immediately
		for _, regex := range regexes {
			if !regex.MatchString(op.parameters.Command) {
				return fmt.Errorf("command does not match the format: %s", regex.String())
			}
		}
	}

	return nil
}

// SetParameters implements [Operation].
func (op *CmdOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	if err := mapstructure.Decode(op.rawParameters, &op.parameters); err != nil {
		return fmt.Errorf("failed to parse parameters for '%s': %w", op.parameters.Executable, err)
	}

	if len(op.parameters.CommandSegments) > 1 && op.parameters.Command != "" {
		return fmt.Errorf("conflicted arguments: can only be arguments or arguments_as_string present")
	}

	if op.parameters.Command != "" {
		op.parameters.CommandSegments = strings.Fields(op.parameters.Command)
	} else {
		op.parameters.Command = strings.Join(op.parameters.CommandSegments, " ")
	}

	if len(op.parameters.CommandSegments) < 1 {
		return fmt.Errorf("not enough segments in shell command")
	}

	op.parameters.Executable = op.parameters.CommandSegments[0]
	if len(op.parameters.CommandSegments) > 1 {
		op.parameters.Arguments = op.parameters.CommandSegments[1:]
	}

	return nil
}

// Run implements [Operation].
func (op *CmdOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	cmdInput := shellexecutors.CommandInput{
		Mode:   shellexecutors.ModeSync,
		Stdout: env.Journaler,
		Stderr: env.Journaler,
		Command: shellexecutors.Command{
			Executable: op.parameters.Executable,
			Arguments:  op.parameters.Arguments,
		},
	}
	if op.parameters.LogCommandInStdStreams {
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

func NewCmdOperation() *CmdOperation {
	return &CmdOperation{}
}
