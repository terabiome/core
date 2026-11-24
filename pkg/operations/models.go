package operations

import (
	"io"

	shellexecutors "github.com/terabiome/core/pkg/executors/shell"
)

type NodeInfo struct {
	MachineID string `json:"machine_id"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}

type OperationEnvironment struct {
	NodeInfo  NodeInfo
	Journaler io.Writer
	Executor  *shellexecutors.CommandExecutor
}
