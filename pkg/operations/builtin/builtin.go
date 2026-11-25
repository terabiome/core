package builtin

import (
	"fmt"

	"github.com/terabiome/core/pkg/operations"
	k3soperations "github.com/terabiome/core/pkg/operations/k3s"
	kerneloperations "github.com/terabiome/core/pkg/operations/kernel"
	shelloperations "github.com/terabiome/core/pkg/operations/shell"
	systemdoperations "github.com/terabiome/core/pkg/operations/systemd"
)

type registration struct {
	name    string
	factory operations.Factory
}

var registrations = []registration{
	{shelloperations.Name, shelloperations.NewCmdOperation},
	{kerneloperations.NameModifyModulesConfig, kerneloperations.NewModifyModulesConfigOperation},
	{kerneloperations.NameModifyParametersConfig, kerneloperations.NewModifyParametersConfigOperation},
	{systemdoperations.NameConfigureUnit, systemdoperations.NewConfigureUnitOperation},
	{k3soperations.NameModifyConfig, k3soperations.NewModifyConfigOperation},
	{k3soperations.NameInstallBinary, k3soperations.NewInstallBinaryOperation},
}

func Register(registry *operations.Registry) error {
	for _, reg := range registrations {
		if err := registry.Register(reg.name, reg.factory); err != nil {
			return fmt.Errorf("failed to register built-in operation %q: %w", reg.name, err)
		}
	}
	return nil
}

func NewRegistry() (*operations.Registry, error) {
	registry := operations.NewRegistry()
	if err := Register(registry); err != nil {
		return nil, err
	}
	return registry, nil
}
