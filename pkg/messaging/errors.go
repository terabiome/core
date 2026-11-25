package messaging

import "errors"

type terminalError struct {
	err error
}

func (e *terminalError) Error() string {
	return e.err.Error()
}

func (e *terminalError) Unwrap() error {
	return e.err
}

// Terminal marks a processing error as non-retryable. Transports acknowledge
// such a message instead of redelivering it.
func Terminal(err error) error {
	if err == nil {
		return nil
	}
	return &terminalError{err: err}
}

func IsTerminal(err error) bool {
	var target *terminalError
	return errors.As(err, &target)
}
