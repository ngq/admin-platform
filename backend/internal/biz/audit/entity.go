package audit

import "time"

// OperationLog 操作日志领域实体
type OperationLog struct {
	ID             int64     `json:"id"`              // 操作日志唯一标识
	UserID         int64     `json:"user_id"`         // 操作用户ID
	Username       string    `json:"username"`        // 操作用户名
	Method         string    `json:"method"`          // HTTP请求方法
	Path           string    `json:"path"`            // 请求路径
	IP             string    `json:"ip"`              // 请求IP地址
	UserAgent      string    `json:"user_agent"`      // 用户代理
	RequestBody    string    `json:"request_body"`    // 请求体内容
	ResponseStatus int       `json:"response_status"` // 响应状态码
	LatencyMs      int       `json:"latency_ms"`      // 请求耗时（毫秒）
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
}

// LoginLog 登录日志领域实体
type LoginLog struct {
	ID        int64     `json:"id"`         // 登录日志唯一标识
	UserID    int64     `json:"user_id"`    // 登录用户ID
	Username  string    `json:"username"`   // 登录用户名
	IP        string    `json:"ip"`         // 登录IP地址
	Location  string    `json:"location"`   // 登录地理位置
	UserAgent string    `json:"user_agent"` // 用户代理
	Status    int8      `json:"status"`     // 登录状态：1-成功 2-失败
	Message   string    `json:"message"`    // 登录结果消息
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// 登录状态常量
const (
	LoginStatusSuccess int8 = 1 // 登录成功
	LoginStatusFail    int8 = 2 // 登录失败
)

// ErrorLog 错误日志领域实体
type ErrorLog struct {
	ID           int64     `json:"id"`            // 错误日志唯一标识
	UserID       int64     `json:"user_id"`       // 操作用户ID
	Username     string    `json:"username"`      // 操作用户名
	Method       string    `json:"method"`        // HTTP请求方法
	Path         string    `json:"path"`          // 请求路径
	ErrorMessage string    `json:"error_message"` // 错误消息
	Stack        string    `json:"stack"`         // 错误堆栈
	Status       int8      `json:"status"`        // 处理状态：1-未处理 2-已处理
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}
