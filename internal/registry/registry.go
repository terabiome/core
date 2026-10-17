package registry

import (
	"fmt"

	"github.com/terabiome/terabiome/internal/operations"
)

var registry = map[string]func() operations.Operation{
	string(operations.OperationNameSystemctlShell): func() operations.Operation {
		return operations.NewSystemctlOperation()
	},
	string(operations.OperationNameFirewallCmdShell): func() operations.Operation {
		return operations.NewFirewallCmdOperation()
	},
	string(operations.OperationNameSysctlShell): func() operations.Operation {
		return operations.NewSysctlOperation()
	},
	string(operations.OperationNameModprobeShell): func() operations.Operation {
		return operations.NewModprobeOperation()
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
