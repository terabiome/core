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

type ModifyParametersConfigOperation struct {
	rawParameters map[string]any
	parameters    struct {
		Entries []sysctlConfEntry `mapstructure:"entries"`
	}
}

func (op *ModifyParametersConfigOperation) Name() string {
	return string(operations.OperationNameKernelModifyParametersConfig)
}

func (op *ModifyParametersConfigOperation) PreCheck(env *operations.OperationEnvironment) error {
	keyPattern := regexp.MustCompile(`^[a-z][a-z0-9_.-]+$`)
	pathPattern := regexp.MustCompile(`^/etc/sysctl.d/[a-zA-Z0-9_\-]+\.conf$`)

	for _, entry := range op.parameters.Entries {
		if !keyPattern.MatchString(entry.Key) {
			return fmt.Errorf("precheck failed: invalid sysctl key '%s'", entry.Key)
		}
		if !pathPattern.MatchString(entry.Path) {
			return fmt.Errorf("precheck failed: invalid sysctl path '%s'", entry.Path)
		}
	}
	return nil
}

func (op *ModifyParametersConfigOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	// 1. Aggregate entries by file path
	fileParams := make(map[string]map[string]string)

	for _, entry := range op.parameters.Entries {
		dir := filepath.Dir(entry.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}

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
		fmt.Fprintf(env.Journaler, "[INFO] Updated sysctl config in %s\n", path)
	}

	return nil
}

func (op *ModifyParametersConfigOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewModifyParametersConfigOperation() operations.Operation {
	return &ModifyParametersConfigOperation{}
}
