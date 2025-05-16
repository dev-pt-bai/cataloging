package errors

import (
	"fmt"
	"slices"

	"github.com/go-sql-driver/mysql"
)

type ErrorCode string

func (e ErrorCode) String() string {
	return string(e)
}

const (
	JSONDecodeFailure            ErrorCode = "400001"
	JSONEncodeFailure            ErrorCode = "400002"
	JSONValidationFailure        ErrorCode = "400003"
	InvalidQueryParameter        ErrorCode = "400004"
	InvalidItemNumberPerPage     ErrorCode = "400005"
	InvalidPageNumber            ErrorCode = "400006"
	UnknownField                 ErrorCode = "400007"
	ParsingFileFailure           ErrorCode = "400008"
	FileOversize                 ErrorCode = "400009"
	UserPasswordMismatch         ErrorCode = "401001"
	MissingAuthorizationHeader   ErrorCode = "401002"
	InvalidAuthorizationType     ErrorCode = "401003"
	InvalidJWTSigningMethod      ErrorCode = "401004"
	ParseTokenFailure            ErrorCode = "401005"
	InvalidToken                 ErrorCode = "401006"
	ExpiredToken                 ErrorCode = "401007"
	InvalidMSGraphAuthCode       ErrorCode = "401008"
	InvalidMSGraphToken          ErrorCode = "401009"
	ResourceIsForbidden          ErrorCode = "403001"
	IllegalUseOfRefreshToken     ErrorCode = "403002"
	IllegalUserOfAccessToken     ErrorCode = "403003"
	ExpiredOTP                   ErrorCode = "403004"
	UserIsUnverified             ErrorCode = "404005"
	UserNotFound                 ErrorCode = "404001"
	UserOTPNotFound              ErrorCode = "404002"
	MaterialTypeNotFound         ErrorCode = "404003"
	MaterialUoMNotFound          ErrorCode = "404004"
	MaterialGroupNotFound        ErrorCode = "404005"
	MaterialPropertiesNotFound   ErrorCode = "404006"
	AssetNotFound                ErrorCode = "404007"
	RequestNotFound              ErrorCode = "404008"
	UserAlreadyExists            ErrorCode = "409001"
	UserOTPAlreadyExists         ErrorCode = "409002"
	UserAlreadyVerified          ErrorCode = "409003"
	MaterialTypeAlreadyExists    ErrorCode = "409004"
	MaterialUoMAlreadyExists     ErrorCode = "409005"
	MaterialGroupAlreadyExists   ErrorCode = "409006"
	AssetAlreadyExists           ErrorCode = "409007"
	UnsupportedFileType          ErrorCode = "415001"
	MissingMSGraphParameter      ErrorCode = "422001"
	MissingMSGraphAuthCode       ErrorCode = "422002"
	MalformedRequestID           ErrorCode = "422003"
	GeneratePasswordFailure      ErrorCode = "500001"
	RunQueryFailure              ErrorCode = "500002"
	RowsAffectedFailure          ErrorCode = "500003"
	ScanRowsFailure              ErrorCode = "500004"
	PrepareStatementFailure      ErrorCode = "500005"
	BuildQueryFailure            ErrorCode = "500006"
	StartingTransactionFailure   ErrorCode = "500007"
	CommittingTransactionFailure ErrorCode = "500008"
	SigningJWTFailure            ErrorCode = "500009"
	CreateHTTPRequestFailure     ErrorCode = "500010"
	SendHTTPRequestFailure       ErrorCode = "500011"
	UndefinedJWTSecret           ErrorCode = "500012"
	GenerateOTPFailure           ErrorCode = "500013"
	CreateFormFileFailure        ErrorCode = "500014"
	CopyFileFailure              ErrorCode = "500015"
	GetMSGraphTokenFailure       ErrorCode = "502001"
	SendEmailFailure             ErrorCode = "502002"
	UploadFileFailure            ErrorCode = "502003"
	DeleteFileFailure            ErrorCode = "502004"
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

func (e *Error) ContainsCodes(codes ...ErrorCode) bool {
	return e != nil && len(codes) > 0 && slices.Contains(codes, e.code)
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

func HasMySQLErrCode(err error, c uint16) bool {
	if merr, ok := err.(*mysql.MySQLError); ok {
		return merr.Number == c
	}
	return false
}
