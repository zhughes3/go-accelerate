package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type StatusCoder interface {
	StatusCode() int
}

type InvalidInputError string

func (e InvalidInputError) StatusCode() int {
	return http.StatusBadRequest
}

func (e InvalidInputError) Error() string {
	return string(e)
}

func NewInvalidInputError(message string) InvalidInputError {
	return InvalidInputError(message)
}

func NewInvalidInputErrorf(format string, args ...any) InvalidInputError {
	return InvalidInputError(fmt.Sprintf(format, args...))
}

func HasInvalidInputError(err error) bool {
	iie := InvalidInputError("")
	return errors.As(err, &iie)
}
