package main

import (
	"fmt"
	"os"

	"github.com/ngq/admin-platform/internal/data/po"
	gorp "github.com/ngq/gorp"
)

func main() {
	if err := gorp.Run(gorp.WithMonoGovernance(), gorp.WithContractHTTPMode(), gorp.WithMigrate(migrate), gorp.WithSetup(setup)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func migrate(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return nil
	}
	if rt.Config.GetString("app.env") == "production" {
		return nil
	}
	rt.DB.Config.DisableForeignKeyConstraintWhenMigrating = true
	return rt.DB.AutoMigrate(&po.AdminUserPO{}, &po.DepartmentPO{}, &po.RolePO{}, &po.UserRolePO{}, &po.MenuPO{}, &po.ButtonPO{}, &po.APIPO{}, &po.RoleMenuPO{}, &po.RoleAPIPO{}, &po.RoleButtonPO{}, &po.SysParamPO{}, &po.OperationLogPO{}, &po.LoginLogPO{}, &po.ErrorLogPO{})
}
