package registry

import (
	"github.com/terabiome/core/pkg/operations"
	k3soperations "github.com/terabiome/core/pkg/operations/k3s"
	kerneloperations "github.com/terabiome/core/pkg/operations/kernel"
	shelloperations "github.com/terabiome/core/pkg/operations/shell"
	systemdoperations "github.com/terabiome/core/pkg/operations/systemd"
)

func NewDefault() *operations.Registry {
	registry := operations.NewRegistry()

	register := func(name string, factory operations.Factory) {
		if err := registry.Register(name, factory); err != nil {
			panic(err)
		}
	}

	register(shelloperations.Name, shelloperations.NewCmdOperation)
	register(kerneloperations.NameModifyModulesConfig, kerneloperations.NewModifyModulesConfigOperation)
	register(kerneloperations.NameModifyParametersConfig, kerneloperations.NewModifyParametersConfigOperation)
	register(systemdoperations.NameConfigureUnit, systemdoperations.NewConfigureUnitOperation)
	register(k3soperations.NameModifyConfig, k3soperations.NewModifyConfigOperation)
	register(k3soperations.NameInstallBinary, k3soperations.NewInstallBinaryOperation)

	return registry
}
