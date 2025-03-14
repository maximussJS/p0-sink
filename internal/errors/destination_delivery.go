package errors

import (
	"errors"
)

type DestinationDeliveryError struct {
	Err error
}

func NewDestinationDeliveryError(msg string) *DestinationDeliveryError {
	err := errors.New(msg)
	return &DestinationDeliveryError{
		Err: err,
	}
}

func (e *DestinationDeliveryError) Error() string {
	return e.Err.Error()
}
