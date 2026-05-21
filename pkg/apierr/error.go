package apierr

import (
	"errors"
	"fmt"
)

// Error 是 proapi 内部统一错误类型。
type Error struct {
	Code       Code           `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	HTTPStatus int            `json:"-"`
	wrapped    error
}

// New 构造一个 Error,根据 code 自动选择 HTTP status。
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message, HTTPStatus: httpStatusByCode(code)}
}

// Wrap 包装一个底层错误并保留 cause 链。
func Wrap(code Code, message string, cause error) *Error {
	e := New(code, message)
	e.wrapped = cause
	return e
}

// WithDetails 附加上下文键值并返回 self,便于链式调用。
func (e *Error) WithDetails(d map[string]any) *Error {
	if e.Details == nil {
		e.Details = map[string]any{}
	}
	for k, v := range d {
		e.Details[k] = v
	}
	return e
}

// Error 返回可读字符串。
func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", int(e.Code), e.Message)
}

// Unwrap 暴露底层错误。
func (e *Error) Unwrap() error { return e.wrapped }

// Is 用于 errors.Is:相同 Code 视为相等。
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}
