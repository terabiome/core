package kernel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/terabiome/internal/operations"
)

type ModifyModulesConfigOperation struct {
	rawParameters map[string]any
	parameters    struct {
		Entries []kernelModuleConfEntry `mapstructure:"entries"`
	}
}

func (op *ModifyModulesConfigOperation) Name() string {
	return string(operations.OperationNameKernelModifyModulesConfig)
}

func (op *ModifyModulesConfigOperation) PreCheck(env *operations.OperationEnvironment) error {
	pathPattern := regexp.MustCompile(`^/etc/modules-load.d/[a-zA-Z0-9_\-]+\.conf$`)
	for _, entry := range op.parameters.Entries {
		if !pathPattern.MatchString(entry.Path) {
			return fmt.Errorf("precheck failed: invalid module path '%s'", entry.Path)
		}
	}
	return nil
}

func (op *ModifyModulesConfigOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	// 1. Aggregate entries by file path
	fileModules := make(map[string]map[string]bool)

	for _, entry := range op.parameters.Entries {
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

func (op *ModifyModulesConfigOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewModifyModulesConfigOperation() *ModifyModulesConfigOperation {
	return &ModifyModulesConfigOperation{}
}
