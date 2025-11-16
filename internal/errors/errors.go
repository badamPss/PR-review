package errors

import (
	"errors"
	"fmt"
)

var (
	NotFoundError      = errors.New("not found")
	AlreadyExistsError = errors.New("already exists")
	BusinessLogicError = errors.New("business logic error")
)

func NewBusinessLogicError(msg string) error {
	return fmt.Errorf("%w: %s", BusinessLogicError, msg)
}

func NewNotFoundError(msg string) error {
	return fmt.Errorf("%w: %s", NotFoundError, msg)
}

func NewAlreadyExistsError(msg string) error {
	return fmt.Errorf("%w: %s", AlreadyExistsError, msg)
}

