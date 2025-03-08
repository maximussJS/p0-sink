package errors

import "errors"

type InvalidBlockOrderError struct {
	Err error
}

func NewInvalidBlockOrderError(msg string) InvalidBlockOrderError {
	err := errors.New(msg)
	return InvalidBlockOrderError{
		Err: err,
	}
}

func (e InvalidBlockOrderError) Error() string {
	return e.Err.Error()
}
