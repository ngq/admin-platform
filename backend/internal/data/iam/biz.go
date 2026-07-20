package iam

import (
	"github.com/ngq/admin-platform/internal/biz/iam"
	"github.com/ngq/admin-platform/internal/biz/system"
	"github.com/ngq/admin-platform/internal/datainfra"
)

// NewBiz 创建 IAM 领域子聚合（DI 桥接）
func NewBiz(d *datainfra.Data, sysParamRepo system.SysParamRepository) *iam.Biz {
	return iam.NewBiz(
		NewUserRepo(d),
		NewRoleRepo(d),
		NewMenuRepo(d),
		NewAPIRepo(d),
		NewButtonRepo(d),
		sysParamRepo,
		NewDepartmentRepo(d),
	)
}
