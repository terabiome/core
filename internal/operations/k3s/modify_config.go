package k3s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/terabiome/internal/operations"
	"go.yaml.in/yaml/v4"
)

type ModifyConfigOperation struct {
	rawParameters map[string]any
	parameters    struct {
		Entries []k3sConfigEntry `mapstructure:"entries"`
	}
}

func (op *ModifyConfigOperation) Name() string {
	return string(operations.OperationNameK3sModifyConfig)
}

func (op *ModifyConfigOperation) PreCheck(env *operations.OperationEnvironment) error {
	for _, entry := range op.parameters.Entries {
		if entry.Path != defaultK3sConfigPath {
			return fmt.Errorf("precheck failed: only %s is supported", defaultK3sConfigPath)
		}
	}
	return nil
}

func (op *ModifyConfigOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	for _, entry := range op.parameters.Entries {
		dir := filepath.Dir(entry.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}

		yamlData, err := yaml.Marshal(entry.Config)
		if err != nil {
			return fmt.Errorf("failed to marshal k3s config: %w", err)
		}

		// Atomic Write
		if err := os.WriteFile(entry.Path, yamlData, 0644); err != nil {
			return fmt.Errorf("failed to write k3s config %s: %w", entry.Path, err)
		}
		fmt.Fprintf(env.Journaler, "[INFO] Updated K3s config in %s\n", entry.Path)
	}
	return nil
}

func (op *ModifyConfigOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewModifyConfigOperation() operations.Operation {
	return &ModifyConfigOperation{}
}
