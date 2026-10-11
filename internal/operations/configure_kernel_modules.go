// take inspirations from old executioner code (v1) lol
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

type kernelModuleParameters struct {
	Immediate bool                `mapstructure:"immediate"`
	Entries   []kernelModuleEntry `mapstructure:"entries"`
}

type kernelModuleEntry struct {
	Path   string `mapstructure:"path"`
	Value  string `mapstructure:"value"`
	Action string `mapstructure:"action"`
}

type ConfigureKernelModulesOperation struct {
	rawParameters map[string]any
	parameters    kernelModuleParameters
}

func (k *ConfigureKernelModulesOperation) Name() string {
	return string(OperationNameConfigureKernelModules)
}

func (k *ConfigureKernelModulesOperation) PreCheck(env *OperationEnvironment) error {
	pathPattern := regexp.MustCompile(`^/etc/modules-load.d/[a-zA-Z0-9_\-]+\.conf$`)
	for _, entry := range k.parameters.Entries {
		if !pathPattern.MatchString(entry.Path) {
			return fmt.Errorf("precheck failed: invalid module path '%s'", entry.Path)
		}
	}
	return nil
}

func (k *ConfigureKernelModulesOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	// 1. Aggregate entries by file path
	fileModules := make(map[string]map[string]bool)

	for _, entry := range k.parameters.Entries {
		if _, ok := fileModules[entry.Path]; !ok {
			fileModules[entry.Path] = make(map[string]bool)
		}

		// Read existing content if file exists
		if len(fileModules[entry.Path]) == 0 {
			content, err := os.ReadFile(entry.Path)
			if err == nil {
				for _, line := range strings.Split(string(content), "\n") {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" {
						fileModules[entry.Path][trimmed] = true
					}
				}
			}
		}

		// Apply action
		switch entry.Action {
		case "add":
			fileModules[entry.Path][entry.Value] = true
		case "remove":
			delete(fileModules[entry.Path], entry.Value)
		}
	}

	// 2. Atomic Write
	for path, modules := range fileModules {
		var lines []string
		for mod := range modules {
			lines = append(lines, mod)
		}
		// Sort for deterministic output (optional but good practice)
		// slices.Sort(lines)

		content := strings.Join(lines, "\n")
		if content != "" {
			content += "\n"
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write module file %s: %w", path, err)
		}
		fmt.Fprintf(env.Journaler, "Updated kernel modules in %s\n", path)
	}

	// 3. Immediate Activation
	if k.parameters.Immediate {
		var allModules []string
		for _, mods := range fileModules {
			for mod := range mods {
				allModules = append(allModules, mod)
			}
		}

		if len(allModules) > 0 {
			cmdInput := shellexecutors.CommandInput{
				Mode:   shellexecutors.ModeSync,
				Stdout: env.Journaler,
				Stderr: env.Journaler,
				Command: shellexecutors.Command{
					Executable: "modprobe",
					Arguments:  append([]string{"-a"}, allModules...),
				},
			}
			output := env.Executor.Execute(ctx, &cmdInput)
			output.Wait()
			if output.Error != nil {
				return fmt.Errorf("failed to activate modules: %w", output.Error)
			}
		}
	}

	return nil
}

func (k *ConfigureKernelModulesOperation) SetParameters(parameters map[string]any) error {
	k.rawParameters = parameters
	return mapstructure.Decode(k.rawParameters, &k.parameters)
}

func NewConfigureKernelModulesOperation() *ConfigureKernelModulesOperation {
	return &ConfigureKernelModulesOperation{}
}
