// Package middleware - 操作日志中间件
//
// 中文说明：
// - 记录 API 请求的操作日志（操作人、路径、方法、IP、请求体、响应状态、耗时等）；
// - 默认只记录写操作（POST/PUT/DELETE）；
// - 可通过配置自定义记录规则；
// - 中间件在请求完成后异步记录日志，不阻塞主流程。
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	bizAudit "github.com/ngq/admin-platform/internal/biz/audit"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/gorp"
)

// OperationLogConfig 操作日志中间件配置
type OperationLogConfig struct {
	// SkipPaths 不记录日志的路径
	SkipPaths []string
	// SkipMethods 不记录日志的 HTTP 方法
	SkipMethods []string
	// RecordBody 是否记录请求体
	RecordBody bool
	// MaxBodySize 最大请求体记录大小（字节）
	MaxBodySize int64
}

// DefaultOperationLogConfig 默认操作日志配置
// 默认只记录写操作（POST/PUT/DELETE），并记录脱敏后的请求体
func DefaultOperationLogConfig() *OperationLogConfig {
	return &OperationLogConfig{
		SkipMethods: []string{"GET", "OPTIONS", "HEAD"},
		RecordBody:  true,
		MaxBodySize: 4096,
	}
}

// OperationLog 创建操作日志中间件
//
// 中文说明：
// - svc 为操作日志服务，用于写入日志记录；
// - config 为配置，控制记录规则；
// - 中间件在请求完成后异步记录日志，不阻塞主流程。
func OperationLog(svc *bizAudit.Biz, config *OperationLogConfig) gorp.Middleware {
	if config == nil {
		config = DefaultOperationLogConfig()
	}

	return func(next gorp.Handler) gorp.Handler {
		return func(ctx gorp.Context) {
			// 检查是否跳过当前方法
			method := ""
			if ctx.Request() != nil {
				method = ctx.Request().Method
			}

			skipMethod := false
			for _, m := range config.SkipMethods {
				if m == method {
					skipMethod = true
					break
				}
			}

			// 检查是否跳过当前路径
			path := ""
			if ctx.Request() != nil {
				path = ctx.Request().URL.Path
			}
			for _, p := range config.SkipPaths {
				if p == path {
					skipMethod = true
					break
				}
			}

			// 记录请求开始时间
			startTime := time.Now()

			// 可选：读取请求体（需要在执行 handler 之前读取，否则会被消费）
			var requestBody string
			if config.RecordBody && !skipMethod && ctx.Request() != nil && ctx.Request().Body != nil {
				bodyBytes, _ := io.ReadAll(io.LimitReader(ctx.Request().Body, config.MaxBodySize))
				requestBody = sanitizeOperationLogBody(bodyBytes)
				// 恢复请求体，让后续 handler 可以继续读取
				ctx.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// 执行请求
			if next != nil {
				next(ctx)
			}

			// 如果是跳过的方法或路径，不记录日志
			if skipMethod {
				return
			}
			if svc == nil {
				return
			}

			// 获取当前用户信息
			var userID int64
			var username string
			if claims, ok := gorp.ClaimsFromContext(ctx); ok {
				userID = claims.SubjectID
				username = claims.SubjectName
			}

			// 获取客户端 IP
			var ip string
			if ctx.Request() != nil {
				// 优先从 X-Forwarded-For / X-Real-IP 获取真实 IP
				ip = ctx.GetHeader("X-Real-IP")
				if ip == "" {
					ip = ctx.GetHeader("X-Forwarded-For")
				}
				if ip == "" && ctx.Request().RemoteAddr != "" {
					// RemoteAddr 格式为 ip:port，需要去掉端口
					idx := len(ctx.Request().RemoteAddr) - 1
					for idx >= 0 && ctx.Request().RemoteAddr[idx] != ':' {
						idx--
					}
					if idx >= 0 {
						ip = ctx.Request().RemoteAddr[:idx]
					} else {
						ip = ctx.Request().RemoteAddr
					}
				}
			}

			// 获取 User-Agent
			var userAgent string
			if ctx.Request() != nil {
				userAgent = ctx.Request().UserAgent()
			}

			// 获取响应状态码（从 RouteContext.ResponseStatus() 获取）
			responseStatus := 0
			if ctx.Response() != nil {
				// 尝试从 ResponseStatus 获取
				if status := ctx.ResponseStatus(); status > 0 {
					responseStatus = status
				}
			}

			// 计算请求耗时
			latencyMs := int(time.Since(startTime).Milliseconds())

			// 构建操作日志
			log := &audit.OperationLog{
				UserID:         userID,
				Username:       username,
				Method:         method,
				Path:           path,
				IP:             ip,
				UserAgent:      userAgent,
				RequestBody:    requestBody,
				ResponseStatus: responseStatus,
				LatencyMs:      latencyMs,
				CreatedAt:      time.Now(),
			}

			// 异步写入日志，不阻塞响应
			// 使用 context.Background() 避免 nil context 导致 GORM panic
			go func() {
				_ = svc.CreateOperationLog(context.Background(), log)
			}()
		}
	}
}

func sanitizeOperationLogBody(body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return ""
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err == nil {
		sanitized := sanitizeOperationLogValue(payload)
		result, err := json.Marshal(sanitized)
		if err == nil {
			return string(result)
		}
		return raw
	}

	values, err := url.ParseQuery(raw)
	if err == nil && len(values) > 0 {
		for key := range values {
			if isSensitiveOperationLogKey(key) {
				values.Set(key, "***")
			}
		}
		return values.Encode()
	}

	return raw
}

func sanitizeOperationLogValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		next := make(map[string]any, len(v))
		for key, item := range v {
			if isSensitiveOperationLogKey(key) {
				next[key] = "***"
				continue
			}
			next[key] = sanitizeOperationLogValue(item)
		}
		return next
	case []any:
		next := make([]any, 0, len(v))
		for _, item := range v {
			next = append(next, sanitizeOperationLogValue(item))
		}
		return next
	default:
		return value
	}
}

func isSensitiveOperationLogKey(key string) bool {
	normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(key))
	sensitiveParts := []string{
		"password",
		"passwd",
		"pwd",
		"token",
		"secret",
		"authorization",
		"credential",
		"captcha",
		"apikey",
	}
	for _, part := range sensitiveParts {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}
