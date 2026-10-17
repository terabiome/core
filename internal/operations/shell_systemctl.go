package operations

type SystemctlOperation struct {
	baseShellOperationOperation
}

// Name implements [Operation].
func (s SystemctlOperation) Name() string {
	return string(OperationNameSystemctlShell)
}

// PreCheck implements [Operation].
func (s SystemctlOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// SetParameters implements [Operation].
func (s *SystemctlOperation) SetParameters(parameters map[string]any) error {
	s.parameters.Executable = "systemctl"
	if err := s.baseShellOperationOperation.SetParameters(parameters); err != nil {
		return err
	}
	return nil
}

func NewSystemctlOperation() *SystemctlOperation {
	return &SystemctlOperation{}
}
