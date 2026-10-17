package registry

import (
	"fmt"

	"github.com/terabiome/terabiome/internal/operations"
)

var registry = map[string]func() operations.Operation{
	string(operations.OperationNameShelllCmd): func() operations.Operation {
		return operations.NewShellCmdOperation()
	},
	string(operations.OperationNameConfigureKernelModulesConf): func() operations.Operation {
		return operations.NewConfigureKernelModulesConfOperation()
	},
	string(operations.OperationNameConfigureKernelParametersConf): func() operations.Operation {
		return operations.NewConfigureKernelParametersConfOperation()
	},
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
