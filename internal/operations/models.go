package operations

type OperationName string

const (
	OperationNameShellCmd OperationName = "shell.cmd"

	OperationNameKernelModifyModulesConfig    OperationName = "kernel.modify-modules-config"
	OperationNameKernelModifyParametersConfig OperationName = "kernel.modify-parameters-config"

	OperationNameK3sInstallBinary OperationName = "k3s.install-binary"
	OperationNameK3sModifyConfig  OperationName = "k3s.modify-config"

	OperationNameSystemdConfigureUnit OperationName = "systemd.configure-unit"
)
