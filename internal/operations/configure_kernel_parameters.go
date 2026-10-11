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

type sysctlParameterParameters struct {
	Immediate bool                   `mapstructure:"immediate"`
	Entries   []sysctlParameterEntry `mapstructure:"entries"`
}

type sysctlParameterEntry struct {
	Path  string `mapstructure:"path"`
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}

type ConfigureKernelParametersOperation struct {
	rawParameters map[string]any
	parameters    sysctlParameterParameters
}

func (s *ConfigureKernelParametersOperation) Name() string {
	return string(OperationNameConfigureKernelParameters)
}

func (s *ConfigureKernelParametersOperation) PreCheck(env *OperationEnvironment) error {
	// Validate sysctl key format (e.g., net.ipv4.ip_forward)
	keyPattern := regexp.MustCompile(`^[a-z][a-z0-9_.-]+$`)
	pathPattern := regexp.MustCompile(`^/etc/sysctl.d/[a-zA-Z0-9_\-]+\.conf$`)

	for _, entry := range s.parameters.Entries {
		if !keyPattern.MatchString(entry.Key) {
			return fmt.Errorf("precheck failed: invalid sysctl key '%s'", entry.Key)
		}
		if !pathPattern.MatchString(entry.Path) {
			return fmt.Errorf("precheck failed: invalid sysctl path '%s'", entry.Path)
		}
	}
	return nil
}

func (s *ConfigureKernelParametersOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	// 1. Aggregate entries by file path
	fileParams := make(map[string]map[string]string)

	for _, entry := range s.parameters.Entries {
		if _, ok := fileParams[entry.Path]; !ok {
			fileParams[entry.Path] = make(map[string]string)

			// Load existing params from disk
			content, err := os.ReadFile(entry.Path)
			if err == nil {
				for _, line := range strings.Split(string(content), "\n") {
					trimmed := strings.TrimSpace(line)
					if trimmed == "" || strings.HasPrefix(trimmed, "#") {
						continue
					}
					parts := strings.SplitN(trimmed, "=", 2)
					if len(parts) == 2 {
						fileParams[entry.Path][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
					}
				}
			}
		}
		// Apply the new value (overwrites if exists)
		fileParams[entry.Path][entry.Key] = entry.Value
	}

	// 2. Atomic Write
	for path, params := range fileParams {
		var lines []string
		for k, v := range params {
			lines = append(lines, fmt.Sprintf("%s = %s", k, v))
		}

		content := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write sysctl file %s: %w", path, err)
		}
		fmt.Fprintf(env.Journaler, "Updated sysctl parameters in %s\n", path)
	}

	// 3. Immediate Application
	if s.parameters.Immediate {
		cmdInput := shellexecutors.CommandInput{
			Mode:   shellexecutors.ModeSync,
			Stdout: env.Journaler,
			Stderr: env.Journaler,
			Command: shellexecutors.Command{
				Executable: "sysctl",
				Arguments:  []string{"--system"},
			},
		}
		output := env.Executor.Execute(ctx, &cmdInput)
		output.Wait()
		if output.Error != nil {
			return fmt.Errorf("failed to apply sysctl changes immediately: %w", output.Error)
		}
	}

	return nil
}

func (s *ConfigureKernelParametersOperation) SetParameters(parameters map[string]any) error {
	s.rawParameters = parameters
	return mapstructure.Decode(s.rawParameters, &s.parameters)
}

func NewConfigureKernelParametersOperation() *ConfigureKernelParametersOperation {
	return &ConfigureKernelParametersOperation{}
}
