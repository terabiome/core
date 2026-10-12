package operations

type FirewallCmdOperation struct {
	baseShellOperationOperation
}

// Name implements [Operation].
func (s FirewallCmdOperation) Name() string {
	return string(OperationNameFirewallCmd)
}

// PreCheck implements [Operation].
func (s FirewallCmdOperation) PreCheck(env *OperationEnvironment) error {
	return nil
}

// SetParameters implements [Operation].
func (s *FirewallCmdOperation) SetParameters(parameters map[string]any) error {
	s.parameters.Executable = "firewall-cmd"
	if err := s.baseShellOperationOperation.SetParameters(parameters); err != nil {
		return err
	}
	return nil
}

func NewFirewallCmdOperation() *FirewallCmdOperation {
	return &FirewallCmdOperation{}
}
