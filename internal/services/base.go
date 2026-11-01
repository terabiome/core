package services

import (
	"fmt"
)

type baseService struct {
}

// shortcut for [fmt.Errorf], with prefix or something
func (s baseService) Errorf(format string, a ...any) error {
	format = fmt.Sprintf("[service] %s", format)
	return fmt.Errorf(format, a)
}
