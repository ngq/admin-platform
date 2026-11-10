package middleware

import (
	stderrors "errors"
	"net/http"

	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

func abortWithError(ctx gorp.Context, err error) {
	code := httpx.CodeInternal
	reason := ""
	message := "internal server error"
	if err == nil {
		code = httpx.CodeSuccess
		message = "success"
	} else {
		var bizErr httpx.BusinessError
		if stderrors.As(err, &bizErr) {
			code = bizErr.Code()
			reason = bizErr.Reason()
			message = bizErr.Message()
		}
	}

	ctx.AbortWithJSON(statusFromErrorCode(code), httpx.Response{
		Code:    code,
		Reason:  reason,
		Message: message,
	})
}

func statusFromErrorCode(code httpx.Code) int {
	switch code {
	case httpx.CodeInvalidArgument, httpx.CodeFailedPrecondition, httpx.CodeOutOfRange:
		return http.StatusBadRequest
	case httpx.CodeUnauthenticated:
		return http.StatusUnauthorized
	case httpx.CodePermissionDenied:
		return http.StatusForbidden
	case httpx.CodeNotFound:
		return http.StatusNotFound
	case httpx.CodeAlreadyExists, httpx.CodeAborted:
		return http.StatusConflict
	case httpx.CodeResourceExhausted:
		return http.StatusTooManyRequests
	case httpx.CodeUnavailable:
		return http.StatusServiceUnavailable
	case httpx.CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
