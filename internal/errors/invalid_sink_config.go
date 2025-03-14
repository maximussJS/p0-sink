package errors

import "errors"

type InvalidSinkConfigError struct {
	Err error
}

func NewInvalidSinkConfigError(msg string) InvalidSinkConfigError {
	err := errors.New(msg)
	return InvalidSinkConfigError{
		Err: err,
	}
}

func (e InvalidSinkConfigError) Error() string {
	return e.Err.Error()
}
