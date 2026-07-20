package iam

import (
	bizIam "github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/errors"
	"github.com/ngq/gorp"
	"github.com/ngq/gorp/framework/httpx"
)

type departmentRequest struct {
	ParentID     int64  `json:"parent_id"`
	Name         string `json:"name" validate:"required,max=64"`
	LeaderUserID int64  `json:"leader_user_id"`
	Sort         int    `json:"sort"`
	Status       int8   `json:"status" validate:"omitempty,oneof=1 2"`
}

type DepartmentHandler struct {
	biz *bizIam.Biz
}

func NewDepartmentHandler(biz *bizIam.Biz) *DepartmentHandler {
	return &DepartmentHandler{biz: biz}
}

// List 获取部门树。
//
// GET /api/v1/departments
func (h *DepartmentHandler) List(ctx gorp.Context) {
	depts, err := h.biz.DepartmentList(ctx.Context(), ctx.Query("name"), ctx.Query("status"))
	if err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, depts)
}

// Get 获取部门详情。
//
// GET /api/v1/departments/:id
func (h *DepartmentHandler) Get(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的部门 ID"))
		return
	}
	dept, err := h.biz.DepartmentGetByID(ctx.Context(), id)
	if err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, dept)
}

// Create 创建部门。
//
// POST /api/v1/departments
func (h *DepartmentHandler) Create(ctx gorp.Context) {
	var req departmentRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}
	dept := departmentFromRequest(req)
	if err := h.biz.DepartmentCreate(ctx.Context(), dept); err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, dept)
}

// Update 更新部门。
//
// POST /api/v1/departments/:id
func (h *DepartmentHandler) Update(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的部门 ID"))
		return
	}
	var req departmentRequest
	if err := ctx.Bind(&req); err != nil {
		httpx.Error(ctx, errors.BadRequest("请求参数错误"))
		return
	}
	if req.ParentID == id {
		httpx.Error(ctx, errors.BadRequest("上级部门不能选择自己"))
		return
	}
	dept := departmentFromRequest(req)
	dept.ID = id
	if err := h.biz.DepartmentUpdate(ctx.Context(), dept); err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, nil)
}

// Delete 删除部门。
//
// POST /api/v1/departments/delete/:id
func (h *DepartmentHandler) Delete(ctx gorp.Context) {
	id, err := ctx.Int64Param("id")
	if err != nil {
		httpx.Error(ctx, errors.BadRequest("无效的部门 ID"))
		return
	}
	if err := h.biz.DepartmentDelete(ctx.Context(), id); err != nil {
		httpx.Error(ctx, err)
		return
	}
	httpx.Success(ctx, nil)
}

func departmentFromRequest(req departmentRequest) *bizIam.Department {
	status := req.Status
	if status == 0 {
		status = bizIam.DepartmentStatusNormal
	}
	return &bizIam.Department{
		ParentID:     req.ParentID,
		Name:         req.Name,
		LeaderUserID: req.LeaderUserID,
		Sort:         req.Sort,
		Status:       status,
	}
}
