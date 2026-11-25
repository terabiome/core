package systemd

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/core/pkg/helpers"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

var unitNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.@-]*$`)

type ConfigureUnitOperation struct {
	rawParameters map[string]any
	parameters    configureSystemdUnitParameters
}

func (op *ConfigureUnitOperation) Name() string {
	return NameConfigureUnit
}

func (op *ConfigureUnitOperation) PreCheck(env *operationspkg.OperationEnvironment) error {
	if op.parameters.Name == "" || op.parameters.Content == "" {
		return fmt.Errorf("precheck failed: 'name' and 'content' are required")
	}
	if !unitNamePattern.MatchString(op.parameters.Name) {
		return fmt.Errorf("precheck failed: invalid unit name %q", op.parameters.Name)
	}
	return nil
}

func (op *ConfigureUnitOperation) Run(ctx context.Context, env *operationspkg.OperationEnvironment) error {
	path := filepath.Join("/etc/systemd/system", op.parameters.Name+".service")
	if err := helpers.File().WriteFileAtomic(path, []byte(op.parameters.Content), 0644); err != nil {
		return fmt.Errorf("failed to write systemd unit %s: %w", path, err)
	}

	fmt.Fprintf(env.Journaler, "Updated systemd unit for %s at %s\n", op.parameters.Name, path)
	return nil
}

func (op *ConfigureUnitOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewConfigureUnitOperation() operationspkg.Operation {
	return &ConfigureUnitOperation{}
}
