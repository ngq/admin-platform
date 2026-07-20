// Package errors defines business-domain error codes and constructors.
// Business error codes are defined here, not in framework layer.
//
// errors 包定义业务域错误码和构造器。
// 业务错误码在此定义，不在框架层定义。
package errors

import (
	"github.com/ngq/gorp/framework/httpx"
)

// ========== 业务错误码 ==========
// Business error codes

const (
	// 通用错误码 / Common error codes
	CodeBadRequest         = 1001
	CodeUnauthorized       = 1002
	CodeForbidden          = 1003
	CodeNotFound           = 1004
	CodeInternalError      = 1005
	CodeServiceUnavailable = 1006
	CodeTooManyRequests    = 1007
	CodeConflict           = 1008
	CodeValidationFailed   = 1009

	// 用户域错误码 / User domain error codes (10001-10999)
	CodeUserNotFound      = 10001
	CodeUserAlreadyExists = 10002
	CodeUserPasswordWrong = 10003
	CodeUserDisabled      = 10004
	CodeUserTokenExpired  = 10005
	CodeUserTokenInvalid  = 10006
)

// ========== 便捷错误构造器 ==========

// New 创建业务错误。
func New(code int, message string) *httpx.BizError {
	return httpx.NewBizError(code, message)
}

// BadRequest 创建参数错误。
func BadRequest(message string) *httpx.BizError {
	return New(CodeBadRequest, message)
}

// Unauthorized 创建未认证错误。
func Unauthorized(message string) *httpx.BizError {
	return New(CodeUnauthorized, message)
}

// Forbidden 创建无权限错误。
func Forbidden(message string) *httpx.BizError {
	return New(CodeForbidden, message)
}

// NotFound 创建资源不存在错误。
func NotFound(message string) *httpx.BizError {
	return New(CodeNotFound, message)
}

// InternalError 创建内部错误。
func InternalError(message string) *httpx.BizError {
	return New(CodeInternalError, message)
}

// Conflict 创建冲突错误。
func Conflict(message string) *httpx.BizError {
	return New(CodeConflict, message)
}

// ========== 用户域便捷构造器 ==========

// UserNotFound 创建用户不存在错误。
func UserNotFound() *httpx.BizError {
	return New(CodeUserNotFound, "用户不存在")
}

// UserAlreadyExists 创建用户已存在错误。
func UserAlreadyExists() *httpx.BizError {
	return New(CodeUserAlreadyExists, "用户名已存在")
}

// UserPasswordWrong 创建密码错误。
func UserPasswordWrong(message string) *httpx.BizError {
	return New(CodeUserPasswordWrong, message)
}

// UserDisabled 创建用户已禁用错误。
func UserDisabled() *httpx.BizError {
	return New(CodeUserDisabled, "用户已被禁用")
}

// UserTokenInvalid 创建令牌无效错误。
func UserTokenInvalid(message string) *httpx.BizError {
	return New(CodeUserTokenInvalid, message)
}
