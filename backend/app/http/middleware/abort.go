package middleware

import (
	stderrors "errors"
	"net/http"

	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

func abortWithError(ctx gorp.Context, err error) {
	code := httpx.CodeInternalError
	message := "internal server error"
	if err == nil {
		code = httpx.CodeSuccess
		message = "success"
	} else {
		var bizErr httpx.BusinessError
		if stderrors.As(err, &bizErr) {
			code = bizErr.Code()
			message = bizErr.Message()
		}
	}

	ctx.AbortWithJSON(statusFromErrorCode(code), httpx.Response{
		Code:    code,
		Message: message,
	})
}

func statusFromErrorCode(code int) int {
	switch code {
	case httpx.CodeBadRequest:
		return http.StatusBadRequest
	case httpx.CodeUnauthorized:
		return http.StatusUnauthorized
	case httpx.CodeForbidden:
		return http.StatusForbidden
	case httpx.CodeNotFound:
		return http.StatusNotFound
	case httpx.CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
