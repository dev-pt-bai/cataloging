package errors

import "fmt"

type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

const (
	JSONDecodeFailure          ErrorCode = "400001"
	JSONValidationFailure      ErrorCode = "400002"
	UserPasswordMismatch       ErrorCode = "401001"
	MissingAuthorizationHeader ErrorCode = "401002"
	InvalidAuthorizationType   ErrorCode = "401003"
	InvalidJWTSigningMethod    ErrorCode = "401004"
	ParseTokenFailure          ErrorCode = "401005"
	InvalidToken               ErrorCode = "401006"
	ExpiredToken               ErrorCode = "401007"
	ResourceIsForbidden        ErrorCode = "403001"
	UserNotFound               ErrorCode = "404001"
	UserAlreadyExists          ErrorCode = "409001"
	GeneratePasswordFailure    ErrorCode = "500001"
	RunQueryFailure            ErrorCode = "500002"
	RowsAffectedFailure        ErrorCode = "500003"
	SigningJWTFailure          ErrorCode = "500004"
)

type Error struct {
	code  ErrorCode
	cause error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.cause == nil {
		return fmt.Sprintf("code: %s", e.code)
	}
	return fmt.Sprintf("code: %s, cause: %s", e.code, e.cause)
}

func (e *Error) Code() string {
	if e == nil {
		return "<nil>"
	}
	return string(e.code)
}

func (e *Error) HasCode(code ErrorCode) bool {
	if e == nil {
		return false
	}
	return e.code == code
}

func (e *Error) Wrap(cause error) *Error {
	if e == nil {
		return nil
	}
	e.cause = cause
	return e
}

func New(code ErrorCode) *Error {
	return &Error{code: code}
}
