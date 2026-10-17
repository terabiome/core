package operations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

type kernelModuleConfEntry struct {
	Path   string `mapstructure:"path"`
	Value  string `mapstructure:"value"`
	Action string `mapstructure:"action"` // "add" or "remove"
}

type ConfigureKernelModulesConfOperation struct {
	rawParameters map[string]any
	parameters    struct {
		Entries []kernelModuleConfEntry `mapstructure:"entries"`
	}
}

func (k *ConfigureKernelModulesConfOperation) Name() string {
	return string(OperationNameConfigureKernelModulesConf)
}

func (k *ConfigureKernelModulesConfOperation) PreCheck(env *OperationEnvironment) error {
	pathPattern := regexp.MustCompile(`^/etc/modules-load.d/[a-zA-Z0-9_\-]+\.conf$`)
	for _, entry := range k.parameters.Entries {
		if !pathPattern.MatchString(entry.Path) {
			return fmt.Errorf("precheck failed: invalid module path '%s'", entry.Path)
		}
	}
	return nil
}

func (k *ConfigureKernelModulesConfOperation) Run(ctx context.Context, env *OperationEnvironment) error {
	// 1. Aggregate entries by file path
	fileModules := make(map[string]map[string]bool)

	for _, entry := range k.parameters.Entries {
		dir := filepath.Dir(entry.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}

		if _, ok := fileModules[entry.Path]; !ok {
			fileModules[entry.Path] = make(map[string]bool)
			// Read existing content if file exists
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

		// Apply action in memory
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

		content := strings.Join(lines, "\n")
		if content != "" {
			content += "\n"
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write module file %s: %w", path, err)
		}
		fmt.Fprintf(env.Journaler, "Updated kernel modules config in %s\n", path)
	}

	return nil
}

func (k *ConfigureKernelModulesConfOperation) SetParameters(parameters map[string]any) error {
	k.rawParameters = parameters
	return mapstructure.Decode(k.rawParameters, &k.parameters)
}

func NewConfigureKernelModulesConfOperation() *ConfigureKernelModulesConfOperation {
	return &ConfigureKernelModulesConfOperation{}
}
