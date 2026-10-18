package registry

import (
	"fmt"

	"github.com/terabiome/terabiome/internal/operations"
	k3soperations "github.com/terabiome/terabiome/internal/operations/k3s"
	kerneloperations "github.com/terabiome/terabiome/internal/operations/kernel"
	shelloperations "github.com/terabiome/terabiome/internal/operations/shell"
	systemdoperations "github.com/terabiome/terabiome/internal/operations/systemd"
)

var registry = OperationRegistry{
	string(operations.OperationNameShellCmd):                     shelloperations.NewCmdOperation,
	string(operations.OperationNameKernelModifyModulesConfig):    kerneloperations.NewModifyModulesConfigOperation,
	string(operations.OperationNameKernelModifyParametersConfig): kerneloperations.NewModifyParametersConfigOperation,
	string(operations.OperationNameSystemdConfigureUnit):         systemdoperations.NewConfigureUnitOperation,
	string(operations.OperationNameK3sModifyConfig):              k3soperations.NewModifyConfigOperation,
	string(operations.OperationNameK3sInstallBinary):             k3soperations.NewInstallBinaryOperation,
}

func Get(name string) (operations.Operation, error) {
	factory, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found in registry", name)
	}
	return factory(), nil
}

func ValidateAll(uow any) error {
	return nil
}
