package operations

type ModprobeOperation struct {
	baseShellOperationOperation
}

// Name implements [Operation].
func (s ModprobeOperation) Name() string {
	return string(OperationNameModprobeShell)
}

// PreCheck implements [Operation].
func (s ModprobeOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// SetParameters implements [Operation].
func (s *ModprobeOperation) SetParameters(parameters map[string]any) error {
	s.parameters.Executable = "modprobe"
	if err := s.baseShellOperationOperation.SetParameters(parameters); err != nil {
		return err
	}
	return nil
}

func NewModprobeOperation() *ModprobeOperation {
	return &ModprobeOperation{}
}
