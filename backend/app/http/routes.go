package http

import (
	"github.com/casbin/casbin/v2"
	hdlraudit "github.com/ngq/admin-platform/app/http/handler/admin/audit"
	hdlriam "github.com/ngq/admin-platform/app/http/handler/admin/iam"
	hdlrsystem "github.com/ngq/admin-platform/app/http/handler/admin/system"
	hdlrsetup "github.com/ngq/admin-platform/app/http/handler/setup"
	"github.com/ngq/admin-platform/app/http/middleware"
	bizaudit "github.com/ngq/admin-platform/internal/biz/audit"
	"github.com/ngq/admin-platform/internal/biz/auth"
	biziam "github.com/ngq/admin-platform/internal/biz/iam"
	bizsystem "github.com/ngq/admin-platform/internal/biz/system"
	svciam "github.com/ngq/admin-platform/internal/service/iam"
	svcsetup "github.com/ngq/admin-platform/internal/service/setup"
	gorp "github.com/ngq/gorp"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

func RegisterRoutes(router gorp.Router, authSvc *svciam.AuthService, roleSvc *svciam.RoleService, iamBiz *biziam.Biz, systemBiz *bizsystem.Biz, auditBiz *bizaudit.Biz, jwtSvc securitycontract.JWTService, enforcer *casbin.SyncedEnforcer, authRepo auth.AuthRepository, resolveMode middleware.APIPermissionModeResolver, setupSvc ...*svcsetup.Service) {
	authH := hdlriam.NewAuthHandler(authSvc)
	userH := hdlriam.NewUserHandler(iamBiz)
	deptH := hdlriam.NewDepartmentHandler(iamBiz)
	roleH := hdlriam.NewRoleHandler(roleSvc)
	menuH := hdlriam.NewMenuHandler(iamBiz)
	apiH := hdlriam.NewAPIHandler(iamBiz)
	buttonH := hdlriam.NewButtonHandler(iamBiz)
	opH := hdlraudit.NewOperationLogHandler(auditBiz)
	loginH := hdlraudit.NewLoginLogHandler(auditBiz)
	errorH := hdlraudit.NewErrorLogHandler(auditBiz)
	paramH := hdlrsystem.NewParamHandler(systemBiz)
	systemH := hdlrsystem.NewSystemHandler(systemBiz)
	authMW := middleware.AuthMiddleware(jwtSvc, "user", authRepo)
	casbinMW := middleware.CasbinRBACWithPermissionMode(enforcer, resolveMode)
	opMW := middleware.OperationLog(auditBiz, nil)
	v1 := router.Group("/api/v1")
	var activeSetupSvc *svcsetup.Service
	if len(setupSvc) > 0 {
		activeSetupSvc = setupSvc[0]
	}
	hdlrsetup.RegisterRoutes(v1, activeSetupSvc)
	public := v1.Group("/auth")
	public.POST("/login", authH.Login)
	public.POST("/refresh", authH.RefreshToken)
	public.GET("/captcha", authH.Captcha)
	a := v1.Group("")
	a.Use(authMW, opMW)
	a.POST("/auth/logout", authH.Logout)
	a.GET("/auth/current", authH.GetCurrentUser)
	a.POST("/auth/password", authH.ChangePassword)
	a.GET("/auth/codes", authH.GetAccessCodes)
	a.GET("/menu/all", menuH.GetAll)
	users := a.Group("/users", casbinMW)
	users.GET("", userH.List)
	users.GET("/:id", userH.Get)
	users.POST("", userH.Create)
	users.POST("/:id", userH.Update)
	users.POST("/delete/:id", userH.Delete)
	users.POST("/password", userH.UpdatePassword)
	users.POST("/roles", userH.AssignRoles)
	depts := a.Group("/departments", casbinMW)
	depts.GET("", deptH.List)
	depts.GET("/:id", deptH.Get)
	depts.POST("", deptH.Create)
	depts.POST("/:id", deptH.Update)
	depts.POST("/delete/:id", deptH.Delete)
	roles := a.Group("/roles", casbinMW)
	roles.GET("", roleH.List)
	roles.GET("/permission-config", roleH.GetPermissionConfig)
	roles.POST("/permission-config", roleH.AssignPermissionConfig)
	roles.GET("/:id", roleH.Get)
	roles.POST("", roleH.Create)
	roles.POST("/:id", roleH.Update)
	roles.POST("/delete/:id", roleH.Delete)
	roles.GET("/menus", roleH.GetMenuIDs)
	roles.POST("/menus", roleH.AssignMenus)
	roles.GET("/apis", roleH.GetAPIIDs)
	roles.POST("/apis", roleH.AssignAPIs)
	roles.GET("/buttons", roleH.GetButtonIDs)
	roles.POST("/buttons", roleH.AssignButtons)
	menus := a.Group("/menus", casbinMW)
	menus.GET("", menuH.List)
	menus.GET("/tree", menuH.GetTree)
	menus.GET("/:id", menuH.Get)
	menus.POST("", menuH.Create)
	menus.POST("/:id", menuH.Update)
	menus.POST("/delete/:id", menuH.Delete)
	menus.GET("/buttons", buttonH.GetByMenuID)
	buttons := a.Group("/buttons", casbinMW)
	buttons.POST("", buttonH.Create)
	buttons.POST("/:id", buttonH.Update)
	buttons.POST("/delete/:id", buttonH.Delete)
	apis := a.Group("/apis", casbinMW)
	apis.GET("", apiH.List)
	apis.GET("/all", apiH.GetAll)
	apis.POST("", apiH.Create)
	apis.POST("/:id", apiH.Update)
	apis.POST("/delete/:id", apiH.Delete)
	apis.POST("/batch-delete", apiH.BatchDelete)
	params := a.Group("/params", casbinMW)
	params.GET("", paramH.List)
	params.GET("/:id", paramH.Get)
	params.POST("", paramH.Create)
	params.POST("/:id", paramH.Update)
	params.POST("/delete/:id", paramH.Delete)
	ops := a.Group("/operation-logs", casbinMW)
	ops.GET("", opH.List)
	ops.POST("/delete/:id", opH.Delete)
	ops.POST("/batch-delete", opH.BatchDelete)
	logins := a.Group("/login-logs", casbinMW)
	logins.GET("", loginH.List)
	logins.POST("/delete/:id", loginH.Delete)
	logins.POST("/batch-delete", loginH.BatchDelete)
	errors := a.Group("/error-logs", casbinMW)
	errors.GET("", errorH.List)
	errors.POST("/:id", errorH.Update)
	errors.POST("/delete/:id", errorH.Delete)
	errors.POST("/batch-delete", errorH.BatchDelete)
	system := a.Group("/system")
	system.GET("/dashboard", systemH.Dashboard)
	system.GET("/info", systemH.SystemInfo)
	system.GET("/health", systemH.HealthCheck)
}
