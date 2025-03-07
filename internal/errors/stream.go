package errors

import "errors"

type StreamTerminationError struct {
	Err error
}

func NewStreamTerminationError(msg string) StreamTerminationError {
	err := errors.New(msg)
	return StreamTerminationError{
		Err: err,
	}
}

func (e StreamTerminationError) Error() string {
	return e.Err.Error()
}
