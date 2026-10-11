package operations

import (
	"github.com/terabiome/terabiome/internal/models"
	"github.com/terabiome/terabiome/internal/writers"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type OperationName string

const (
	OperationNameDemo                      OperationName = "demo"
	OperationNameSystemctl                 OperationName = "systemctl"
	OperationNameFirewallCmd               OperationName = "firewall-cmd"
	OperationNameSysctl                    OperationName = "sysctl"
	OperationNameConfigureKernelModules    OperationName = "configure-kernel-modules"
	OperationNameConfigureKernelParameters OperationName = "configure-kernel-parameters"
)

type OperationEnvironment struct {
	NodeInfo  models.NodeInfo
	Journaler *writers.Journaler
	Executor  *shellexecutors.CommandExecutor
}
