package errors

import (
	"fmt"
	"slices"
)

type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

const (
	JSONDecodeFailure          ErrorCode = "400001"
	JSONValidationFailure      ErrorCode = "400002"
	InvalidQueryParameter      ErrorCode = "400003"
	InvalidItemNumberPerPage   ErrorCode = "400004"
	InvalidPageNumber          ErrorCode = "400005"
	UserPasswordMismatch       ErrorCode = "401001"
	MissingAuthorizationHeader ErrorCode = "401002"
	InvalidAuthorizationType   ErrorCode = "401003"
	InvalidJWTSigningMethod    ErrorCode = "401004"
	ParseTokenFailure          ErrorCode = "401005"
	InvalidToken               ErrorCode = "401006"
	ExpiredToken               ErrorCode = "401007"
	ResourceIsForbidden        ErrorCode = "403001"
	IllegalUseOfRefreshToken   ErrorCode = "403002"
	IllegalUserOfAccessToken   ErrorCode = "403003"
	UserNotFound               ErrorCode = "404001"
	MaterialTypeNotFound       ErrorCode = "404002"
	UserAlreadyExists          ErrorCode = "409001"
	MaterialTypeAlreadyExists  ErrorCode = "409002"
	GeneratePasswordFailure    ErrorCode = "500001"
	RunQueryFailure            ErrorCode = "500002"
	RowsAffectedFailure        ErrorCode = "500003"
	ScanRowsFailure            ErrorCode = "500004"
	BuildQueryFailure          ErrorCode = "500005"
	UnknownField               ErrorCode = "500006"
	SigningJWTFailure          ErrorCode = "500007"
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

func (e *Error) HasCodes(codes ...ErrorCode) bool {
	if e == nil {
		return false
	}
	return slices.Contains(codes, e.code)
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
