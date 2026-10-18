package operations

import (
	"github.com/terabiome/terabiome/internal/models"
	"github.com/terabiome/terabiome/internal/writers"
	shellexecutors "github.com/terabiome/terabiome/pkg/executors/shell"
)

type OperationName string

// Shell namespace
const (
	OperationNameShellCmd OperationName = "shell.cmd"
)

// Kernel namespace
const (
	OperationNameKernelModifyModulesConfig    OperationName = "kernel.modify-modules-config"
	OperationNameKernelModifyParametersConfig OperationName = "kernel.modify-parameters-config"
)

// K3s namespace
const (
	OperationNameK3sInstallBinary OperationName = "k3s.install-binary"
	OperationNameK3sModifyConfig  OperationName = "k3s.modify-config"
)

// Systemd namespace
const (
	OperationNameSystemdConfigureUnit OperationName = "systemd.configure-unit"
)

type OperationEnvironment struct {
	NodeInfo  models.NodeInfo
	Journaler *writers.Journaler
	Executor  *shellexecutors.CommandExecutor
}
