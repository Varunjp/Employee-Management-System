package domain

import "errors"

var (
	// ErrEmployeeNotFound is returned when an employee id has no matching record.
	ErrEmployeeNotFound = errors.New("employee not found")

	// ErrInvalidInput is returned when the caller-supplied data fails validation.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInternal wraps unexpected infrastructure failures (db, cache, etc).
	ErrInternal = errors.New("internal server error")
)
