package service

import "errors"

var (
	ErrBadRequest         = errors.New("bad_request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotFound           = errors.New("not_found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrInternal           = errors.New("internal_error")
)

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewError(code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func BadRequest(message string) *Error {
	return NewError("bad_request", message, ErrBadRequest)
}

func Unauthorized(message string) *Error {
	return NewError("unauthorized", message, ErrUnauthorized)
}

func NotFound(message string) *Error {
	return NewError("not_found", message, ErrNotFound)
}

func Conflict(message string) *Error {
	return NewError("conflict", message, ErrConflict)
}

func Internal(message string, err error) *Error {
	return NewError("internal_error", message, err)
}
