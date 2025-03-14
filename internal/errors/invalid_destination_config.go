package errors

import "errors"

type InvalidDestinationConfigError struct {
	Err error
}

func NewInvalidDestinationConfigError(msg string) InvalidDestinationConfigError {
	err := errors.New(msg)
	return InvalidDestinationConfigError{
		Err: err,
	}
}

func (e InvalidDestinationConfigError) Error() string {
	return e.Err.Error()
}
