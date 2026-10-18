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

type OperationEnvironment struct {
	NodeInfo  models.NodeInfo
	Journaler *writers.Journaler
	Executor  *shellexecutors.CommandExecutor
}
