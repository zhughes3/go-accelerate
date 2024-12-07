package errors

import "fmt"

// Wrap returns an err wrapped in a new error.
// The wrapped error's value will be appended to the supplied message.
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf return an err wrapped in a new error.
// The wrapped error's value will be appended to the result of the format specification.
func Wrapf(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}
