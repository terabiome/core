package registry

import (
	"fmt"

	"github.com/terabiome/terabiome/internal/operations"
	kerneloperations "github.com/terabiome/terabiome/internal/operations/kernel"
	shelloperations "github.com/terabiome/terabiome/internal/operations/shell"
)

var registry = map[string]func() operations.Operation{
	string(operations.OperationNameShellCmd): func() operations.Operation {
		return shelloperations.NewCmdOperation()
	},
	string(operations.OperationNameKernelModifyModulesConfig): func() operations.Operation {
		return kerneloperations.NewModifyModulesConfigOperation()
	},
	string(operations.OperationNameKernelModifyParametersConfig): func() operations.Operation {
		return kerneloperations.NewModifyParametersConfigOperation()
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
