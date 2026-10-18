package systemd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/terabiome/internal/operations"
)

type ConfigureUnitOperation struct {
	rawParameters map[string]any
	parameters    configureSystemdUnitParameters
}

// Name implements [operations.Operation].
func (op *ConfigureUnitOperation) Name() string {
	return string(operations.OperationNameSystemdConfigureUnit)
}

// PreCheck implements [operations.Operation].
func (op *ConfigureUnitOperation) PreCheck(env *operations.OperationEnvironment) error {
	if op.parameters.Name == "" || op.parameters.Content == "" {
		return fmt.Errorf("precheck failed: 'name' and 'content' are required")
	}
	return nil
}

// Run implements [operations.Operation].
func (op *ConfigureUnitOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	path := filepath.Join("/etc/systemd/system", op.parameters.Name+".service")

	// Atomic Write to ensure the unit file is never partially written
	if err := os.WriteFile(path, []byte(op.parameters.Content), 0644); err != nil {
		return fmt.Errorf("failed to write systemd unit %s: %w", path, err)
	}

	fmt.Fprintf(env.Journaler, "Updated systemd unit for %s at %s\n", op.parameters.Name, path)
	return nil
}

// SetParameters implements [operations.Operation].
func (op *ConfigureUnitOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewConfigureUnitOperation() operations.Operation {
	return &ConfigureUnitOperation{}
}
