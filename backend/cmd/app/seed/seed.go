package seed

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/ngq/admin-platform/internal/data/po"
	"github.com/ngq/admin-platform/internal/datainfra"
	svciam "github.com/ngq/admin-platform/internal/service/iam"
	gorp "github.com/ngq/gorp"
	"gorm.io/gorm"
)

func Run(_ *gorp.HTTPRuntime, d *datainfra.Data, enforcer *casbin.SyncedEnforcer) error {
	return d.DB(context.Background()).Transaction(func(tx *gorm.DB) error {
		dept := po.DepartmentPO{}
		if err := tx.Where("name = ?", "总部").FirstOrCreate(&dept, po.DepartmentPO{Name: "总部", Status: 1}).Error; err != nil {
			return err
		}
		role := po.RolePO{}
		if err := tx.Where("code = ?", "super-admin").FirstOrCreate(&role, po.RolePO{Name: "超级管理员", Code: "super-admin", DataScope: 1, Status: 1, Description: "拥有全部权限"}).Error; err != nil {
			return err
		}
		var user po.AdminUserPO
		if err := tx.Where("username = ?", "admin").First(&user).Error; err == gorm.ErrRecordNotFound {
			password, e := svciam.HashPassword("admin123")
			if e != nil {
				return e
			}
			user = po.AdminUserPO{Username: "admin", Password: password, Nickname: "超级管理员", DeptID: dept.ID, Status: 1}
			if e = tx.Create(&user).Error; e != nil {
				return e
			}
		} else if err != nil {
			return err
		}
		_ = tx.FirstOrCreate(&po.UserRolePO{}, po.UserRolePO{UserID: user.ID, RoleID: role.ID}).Error
		menus := []po.MenuPO{{Name: "工作台", Path: "/dashboard", Component: "dashboard/analytics/index", Icon: "lucide:layout-dashboard", Type: 2, Visible: 1, Status: 1}, {Name: "系统管理", Path: "/system", Redirect: "/system/user", Icon: "lucide:settings", Type: 1, Visible: 1, Status: 1}, {Name: "日志管理", Path: "/logs", Redirect: "/logs/operation", Icon: "lucide:scroll-text", Type: 1, Visible: 1, Status: 1}}
		for i := range menus {
			if err := tx.Where("path = ?", menus[i].Path).FirstOrCreate(&menus[i]).Error; err != nil {
				return err
			}
		}
		children := []po.MenuPO{{ParentID: menus[1].ID, Name: "用户管理", Path: "/system/user", Component: "system/user/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[1].ID, Name: "部门管理", Path: "/system/dept", Component: "system/dept/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[1].ID, Name: "角色管理", Path: "/system/role", Component: "system/role/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[1].ID, Name: "菜单管理", Path: "/system/menu", Component: "system/menu/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[1].ID, Name: "API 管理", Path: "/system/api", Component: "system/api/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[1].ID, Name: "系统参数", Path: "/system/param", Component: "system/param/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[2].ID, Name: "操作日志", Path: "/logs/operation", Component: "logs/operation/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[2].ID, Name: "登录日志", Path: "/logs/login", Component: "logs/login/index", Type: 2, Visible: 1, Status: 1}, {ParentID: menus[2].ID, Name: "错误日志", Path: "/logs/error", Component: "logs/error/index", Type: 2, Visible: 1, Status: 1}}
		for i := range children {
			if err := tx.Where("path = ?", children[i].Path).FirstOrCreate(&children[i]).Error; err != nil {
				return err
			}
		}
		buttonDefs := map[string][]po.ButtonPO{"/system/user": {{Name: "创建用户", Code: "user:create"}, {Name: "编辑用户", Code: "user:update"}, {Name: "删除用户", Code: "user:delete"}, {Name: "修改密码", Code: "user:password"}, {Name: "分配角色", Code: "user:roles"}}, "/system/dept": {{Name: "创建部门", Code: "department:create"}, {Name: "编辑部门", Code: "department:update"}, {Name: "删除部门", Code: "department:delete"}}, "/system/role": {{Name: "创建角色", Code: "role:create"}, {Name: "编辑角色", Code: "role:update"}, {Name: "删除角色", Code: "role:delete"}, {Name: "权限配置", Code: "role:permissions"}}, "/system/menu": {{Name: "创建菜单", Code: "menu:create"}, {Name: "编辑菜单", Code: "menu:update"}, {Name: "删除菜单", Code: "menu:delete"}, {Name: "按钮管理", Code: "menu:buttons"}}, "/system/api": {{Name: "创建 API", Code: "api:create"}, {Name: "编辑 API", Code: "api:update"}, {Name: "删除 API", Code: "api:delete"}}}
		for path, defs := range buttonDefs {
			var menu po.MenuPO
			if err := tx.Where("path = ?", path).First(&menu).Error; err != nil {
				return err
			}
			for _, b := range defs {
				b.MenuID = menu.ID
				if err := tx.Where("menu_id = ? AND code = ?", b.MenuID, b.Code).FirstOrCreate(&b).Error; err != nil {
					return err
				}
			}
		}
		var allMenus []po.MenuPO
		var allButtons []po.ButtonPO
		tx.Find(&allMenus)
		tx.Find(&allButtons)
		for _, m := range allMenus {
			_ = tx.FirstOrCreate(&po.RoleMenuPO{}, po.RoleMenuPO{RoleID: role.ID, MenuID: m.ID}).Error
		}
		for _, b := range allButtons {
			_ = tx.FirstOrCreate(&po.RoleButtonPO{}, po.RoleButtonPO{RoleID: role.ID, ButtonID: b.ID}).Error
		}
		params := []po.SysParamPO{{Key: "captcha_enabled", Value: "false", Category: "security", Status: 1}, {Key: "password_min_length", Value: "6", Category: "security", Status: 1}}
		for _, p := range params {
			_ = tx.Where("`key` = ?", p.Key).FirstOrCreate(&p).Error
		}
		if _, err := enforcer.AddPolicy("super-admin", "*", "*"); err != nil {
			return fmt.Errorf("casbin: %w", err)
		}
		return enforcer.SavePolicy()
	})
}
