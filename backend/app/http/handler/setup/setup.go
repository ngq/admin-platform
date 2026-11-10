// Package setup provides HTTP handlers for the database setup wizard.
package setup

import (
	"github.com/ngq/admin-platform/internal/errors"
	svcsetup "github.com/ngq/admin-platform/internal/service/setup"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

// Handler handles setup status, connection testing, and database initialization endpoints.
type Handler struct {
	svc *svcsetup.Service
}

// NewHandler creates a setup HTTP Handler.
func NewHandler(svc *svcsetup.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the setup endpoints onto the given router group (typically /api/v1).
func RegisterRoutes(v1 gorp.Router, svc *svcsetup.Service) {
	if v1 == nil {
		return
	}
	if svc == nil {
		svc = svcsetup.NewService(svcsetup.Options{Initialized: true})
	}
	h := NewHandler(svc)
	group := v1.Group("/setup")
	group.GET("/status", h.Status)
	group.POST("/test-connection", h.TestConnection)
	group.POST("/database", h.InitializeDatabase)
}

// Status returns whether the database is already initialized.
// GET /api/v1/setup/status
func (h *Handler) Status(ctx gorp.Context) {
	httpx.Success(ctx, h.svc.GetStatus())
}

// TestConnection tests database server connectivity without modifying data or config.
// POST /api/v1/setup/test-connection
func (h *Handler) TestConnection(ctx gorp.Context) {
	if h.svc.IsInitialized() {
		httpx.Error(ctx, errors.Forbidden("数据库已完成初始化，禁止重复操作"))
		return
	}

	var req svcsetup.DatabaseInitInput
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请完整填写有效的数据库连接参数"))
		return
	}
	res, err := h.svc.TestConnection(ctx.Context(), req)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, res)
}

// InitializeDatabase creates the database if needed, applies goose migrations, seeds initial data,
// persists configuration to config/app.yaml, and triggers service readiness.
// POST /api/v1/setup/database
func (h *Handler) InitializeDatabase(ctx gorp.Context) {
	if h.svc.IsInitialized() {
		httpx.Error(ctx, errors.Forbidden("数据库已完成初始化，禁止重复初始化"))
		return
	}

	var req svcsetup.DatabaseInitInput
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请完整填写有效的数据库连接参数"))
		return
	}
	res, err := h.svc.InitializeDatabase(ctx.Context(), req)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, res)
}
