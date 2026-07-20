package iam

import (
	"github.com/ngq/admin-platform/app/http/handler/common"
	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

type createAPIRequest struct {
	Group          string `json:"group"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	Description    string `json:"description"`
	MenuID         int64  `json:"menu_id"`
	ActionType     string `json:"action_type"`
	PermissionMode string `json:"permission_mode"`
	Ignored        int8   `json:"ignored"`
}

type updateAPIRequest struct {
	Group          string `json:"group"`
	Path           string `json:"path"`
	Method         string `json:"method"`
	Description    string `json:"description"`
	MenuID         int64  `json:"menu_id"`
	ActionType     string `json:"action_type"`
	PermissionMode string `json:"permission_mode"`
	Ignored        int8   `json:"ignored"`
}

type batchDeleteRequest struct {
	IDs []int64 `json:"ids"`
}

type APIHandler struct {
	biz *bizIam.Biz
}

func NewAPIHandler(biz *bizIam.Biz) *APIHandler {
	return &APIHandler{biz: biz}
}

func (h *APIHandler) List(ctx gorp.Context) {
	pg := common.ParsePagination(ctx, 20, 100)
	group := ctx.Query("group")
	path := ctx.Query("path")
	method := ctx.Query("method")

	apis, total, err := h.biz.APIList(ctx.Context(), pg.Page, pg.PageSize, group, path, method)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.SuccessPaginated(ctx, apis, total, pg.Page, pg.PageSize)
}

func (h *APIHandler) GetAll(ctx gorp.Context) {
	apis, err := h.biz.APIGetAll(ctx.Context())
	if err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, apis)
}

func (h *APIHandler) Create(ctx gorp.Context) {
	var req createAPIRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	api := &bizIam.APIResource{
		Group:          req.Group,
		Path:           req.Path,
		Method:         req.Method,
		Description:    req.Description,
		MenuID:         req.MenuID,
		ActionType:     req.ActionType,
		PermissionMode: req.PermissionMode,
		Ignored:        req.Ignored,
	}

	if err := h.biz.APICreate(ctx.Context(), api); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

func (h *APIHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的 API 资源 ID"))
		return
	}

	var req updateAPIRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	api := &bizIam.APIResource{
		ID:             id,
		Group:          req.Group,
		Path:           req.Path,
		Method:         req.Method,
		Description:    req.Description,
		MenuID:         req.MenuID,
		ActionType:     req.ActionType,
		PermissionMode: req.PermissionMode,
		Ignored:        req.Ignored,
	}

	if err := h.biz.APIUpdate(ctx.Context(), api); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

func (h *APIHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的 API 资源 ID"))
		return
	}

	if err := h.biz.APIDelete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}

func (h *APIHandler) BatchDelete(ctx gorp.Context) {
	var req batchDeleteRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("参数绑定失败: "+err.Error()))
		return
	}

	if err := h.biz.APIBatchDelete(ctx.Context(), req.IDs); err != nil {
		httpx.Error(ctx, err)
		return
	}

	httpx.Success(ctx, nil)
}
