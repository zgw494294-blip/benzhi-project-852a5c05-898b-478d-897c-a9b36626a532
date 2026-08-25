package domain

import "fmt"

// Error 是可稳定映射到协议层的业务错误。
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func NewError(code, message string) error { return &Error{Code: code, Message: message} }

func Invalid(field, reason string) error {
	return &Error{Code: "INVALID_" + field, Message: reason}
}

func Conflict(message string) error { return &Error{Code: "STATE_CONFLICT", Message: message} }

func NotFound(kind, id string) error {
	return &Error{Code: "NOT_FOUND", Message: fmt.Sprintf("%s %s 不存在", kind, id)}
}

func ErrorCode(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return "INTERNAL_ERROR"
}
