package operations

import (
	"github.com/terabiome/terabiome/internal/models"
	"github.com/terabiome/terabiome/internal/writers"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type OperationName string

const (
	OperationNameSystemctlShell                OperationName = "systemctl-shell"
	OperationNameFirewallCmdShell              OperationName = "firewall-cmd-shell"
	OperationNameSysctlShell                   OperationName = "sysctl-shell"
	OperationNameModprobeShell                 OperationName = "modprobe-shell"
	OperationNameConfigureKernelModulesConf    OperationName = "configure-kernel-modules-conf"
	OperationNameConfigureKernelParametersConf OperationName = "configure-kernel-parameters-conf"
)

type OperationEnvironment struct {
	NodeInfo  models.NodeInfo
	Journaler *writers.Journaler
	Executor  *shellexecutors.CommandExecutor
}
