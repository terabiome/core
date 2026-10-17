package operations

type SysctlOperation struct {
	baseShellOperationOperation
}

// Name implements [Operation].
func (s SysctlOperation) Name() string {
	return string(OperationNameSysctlShell)
}

// PreCheck implements [Operation].
func (s SysctlOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// SetParameters implements [Operation].
func (s *SysctlOperation) SetParameters(parameters map[string]any) error {
	s.parameters.Executable = "sysctl"
	if err := s.baseShellOperationOperation.SetParameters(parameters); err != nil {
		return err
	}
	return nil
}

func NewSysctlOperation() *SysctlOperation {
	return &SysctlOperation{}
}
