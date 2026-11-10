// Package routesync synchronizes registered permission routes into sys_apis.
package routesync

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/ngq/admin-platform/internal/data/po"
	"github.com/ngq/admin-platform/internal/datainfra"
	gorp "github.com/ngq/gorp"
	"gorm.io/gorm"
)

// Route is the provider-independent route metadata needed by permission sync.
type Route struct {
	Method string
	Path   string
}

// Registry decorates the gorp Router contract and records routes as they are registered.
type Registry struct {
	mu     sync.RWMutex
	routes []Route
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Wrap(router gorp.Router) gorp.Router {
	return &trackingRouter{Router: router, registry: r}
}

func (r *Registry) Routes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	routes := make([]Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

func (r *Registry) add(method, path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, Route{Method: method, Path: path})
}

type trackingRouter struct {
	gorp.Router
	registry *Registry
	prefix   string
}

func (r *trackingRouter) Use(middleware ...gorp.Middleware) { r.Router.Use(middleware...) }

func (r *trackingRouter) Group(prefix string, middleware ...gorp.Middleware) gorp.Router {
	return &trackingRouter{
		Router:   r.Router.Group(prefix, middleware...),
		registry: r.registry,
		prefix:   joinPath(r.prefix, prefix),
	}
}

func (r *trackingRouter) Handle(method, path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.registry.add(method, joinPath(r.prefix, path))
	r.Router.Handle(method, path, handler, middleware...)
}

func (r *trackingRouter) HandleFunc(method, path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.registry.add(method, joinPath(r.prefix, path))
	r.Router.HandleFunc(method, path, handler, middleware...)
}

func (r *trackingRouter) GET(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodGet, path, handler, middleware...)
}

func (r *trackingRouter) POST(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodPost, path, handler, middleware...)
}

func (r *trackingRouter) PUT(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodPut, path, handler, middleware...)
}

func (r *trackingRouter) DELETE(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodDelete, path, handler, middleware...)
}

func (r *trackingRouter) PATCH(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodPatch, path, handler, middleware...)
}

func (r *trackingRouter) HEAD(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodHead, path, handler, middleware...)
}

func (r *trackingRouter) OPTIONS(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle(http.MethodOptions, path, handler, middleware...)
}

func (r *trackingRouter) ANY(path string, handler gorp.Handler, middleware ...gorp.Middleware) {
	r.Handle("ANY", path, handler, middleware...)
}

func (r *trackingRouter) Mount(path string, handler http.Handler) { r.Router.Mount(path, handler) }

func joinPath(prefix, path string) string {
	if prefix == "" {
		if path == "" {
			return "/"
		}
		return "/" + strings.TrimPrefix(path, "/")
	}
	if path == "" || path == "/" {
		return "/" + strings.Trim(strings.TrimSuffix(prefix, "/")+path, "/")
	}
	return "/" + strings.Trim(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}

var apiMenuRules = []struct {
	prefix   string
	menuPath string
}{
	{"/api/v1/roles/permission-config", "/system/role"},
	{"/api/v1/departments", "/system/dept"},
	{"/api/v1/users", "/system/user"},
	{"/api/v1/roles", "/system/role"},
	{"/api/v1/menus", "/system/menu"},
	{"/api/v1/menu/", "/system/menu"},
	{"/api/v1/buttons", "/system/menu"},
	{"/api/v1/apis", "/system/api"},
	{"/api/v1/params", "/system/param"},
	{"/api/v1/operation-logs", "/logs/operation"},
	{"/api/v1/login-logs", "/logs/login"},
	{"/api/v1/error-logs", "/logs/error"},
}

var groupRules = []struct {
	prefix string
	group  string
}{
	{"/api/v1/auth/", "auth"},
	{"/api/v1/users", "user"},
	{"/api/v1/departments", "department"},
	{"/api/v1/roles", "role"},
	{"/api/v1/menus", "menu"},
	{"/api/v1/menu/", "menu"},
	{"/api/v1/buttons", "button"},
	{"/api/v1/apis", "api"},
	{"/api/v1/params", "param"},
	{"/api/v1/operation-logs", "operation-log"},
	{"/api/v1/login-logs", "login-log"},
	{"/api/v1/error-logs", "error-log"},
	{"/api/v1/system/", "system"},
	{"/api/v1/setup/", "setup"},
}

type apiDescriptionKey struct {
	method string
	path   string
}

var apiDescriptions = map[apiDescriptionKey]string{
	{"GET", "/api/v1/departments"}:              "Department list",
	{"GET", "/api/v1/departments/:id"}:          "Department details",
	{"POST", "/api/v1/departments"}:             "Create department",
	{"POST", "/api/v1/departments/:id"}:         "Update department",
	{"POST", "/api/v1/departments/delete/:id"}:  "Delete department",
	{"GET", "/api/v1/roles/permission-config"}:  "Role permission configuration",
	{"POST", "/api/v1/roles/permission-config"}: "Save role permission configuration",
	{"GET", "/api/v1/setup/status"}:             "Setup status",
	{"POST", "/api/v1/setup/test-connection"}:   "Test database connection",
	{"POST", "/api/v1/setup/database"}:          "Initialize database",
}

var ignoredPaths = map[string]bool{
	"/api/v1/auth/login":            true,
	"/api/v1/auth/refresh":          true,
	"/api/v1/auth/captcha":          true,
	"/api/v1/system/health":         true,
	"/api/v1/setup/status":          true,
	"/api/v1/setup/test-connection": true,
	"/api/v1/setup/database":        true,
}

// Sync imports the currently registered Contract HTTP routes into sys_apis.
func Sync(routes []Route, d *datainfra.Data) error {
	db := d.DB(context.Background())
	menuIDByPath := loadMenuIDByPath(db)

	type routeKey struct{ method, path string }
	liveRoutes := make(map[routeKey]po.APIPO)
	for _, route := range routes {
		if !strings.HasPrefix(route.Path, "/api/v1/") {
			continue
		}
		key := routeKey{route.Method, route.Path}
		ignored := int8(0)
		permissionMode := "casbin"
		if ignoredPaths[route.Path] {
			ignored = 1
			permissionMode = "ignored"
		}
		liveRoutes[key] = po.APIPO{
			Group:          deriveGroup(route.Path),
			Path:           route.Path,
			Method:         route.Method,
			Description:    deriveDescription(route.Path, route.Method),
			MenuID:         deriveMenuID(route.Path, menuIDByPath),
			ActionType:     "other",
			PermissionMode: permissionMode,
			Ignored:        ignored,
		}
	}

	var existing []po.APIPO
	db.Unscoped().Find(&existing)
	dbKeys := make(map[routeKey]po.APIPO, len(existing))
	for _, api := range existing {
		dbKeys[routeKey{api.Method, api.Path}] = api
	}

	var toInsert, toRestore, toUpdateDescription, toUpdateMenuID []po.APIPO
	for key, api := range liveRoutes {
		current, found := dbKeys[key]
		if !found {
			toInsert = append(toInsert, api)
			continue
		}
		if !current.DeletedAt.Time.IsZero() {
			api.ID = current.ID
			toRestore = append(toRestore, api)
			continue
		}
		if strings.TrimSpace(current.Description) == "" && api.Description != "" {
			api.ID = current.ID
			toUpdateDescription = append(toUpdateDescription, api)
		}
		if current.MenuID == 0 && api.MenuID != 0 {
			api.ID = current.ID
			toUpdateMenuID = append(toUpdateMenuID, api)
		}
	}

	var toSoftDelete []int64
	for key, api := range dbKeys {
		if _, found := liveRoutes[key]; !found && api.DeletedAt.Time.IsZero() {
			toSoftDelete = append(toSoftDelete, api.ID)
		}
	}

	if len(toInsert) > 0 {
		if err := db.Create(&toInsert).Error; err != nil {
			return err
		}
	}
	for _, api := range toRestore {
		if err := db.Unscoped().Model(&po.APIPO{}).Where("id = ?", api.ID).Updates(map[string]interface{}{
			"deleted_at": nil, "group_name": api.Group, "description": api.Description,
			"permission_mode": api.PermissionMode, "ignored": api.Ignored,
		}).Error; err != nil {
			return err
		}
	}
	for _, api := range toUpdateDescription {
		if err := db.Model(&po.APIPO{}).Where("id = ?", api.ID).Update("description", api.Description).Error; err != nil {
			return err
		}
	}
	for _, api := range toUpdateMenuID {
		if err := db.Model(&po.APIPO{}).Where("id = ?", api.ID).Update("menu_id", api.MenuID).Error; err != nil {
			return err
		}
	}
	if len(toSoftDelete) > 0 {
		if err := db.Where("id IN ?", toSoftDelete).Delete(&po.APIPO{}).Error; err != nil {
			return err
		}
	}

	syncSuperAdminRoleAPIs(db)
	return nil
}

func loadMenuIDByPath(db *gorm.DB) map[string]int64 {
	var menus []po.MenuPO
	db.Select("id", "path").Where("path <> ''").Find(&menus)
	result := make(map[string]int64, len(menus))
	for _, menu := range menus {
		if menu.Path != "" {
			result[menu.Path] = menu.ID
		}
	}
	return result
}

func deriveMenuPath(path string) string {
	type rule struct{ prefix, menuPath string }
	sorted := make([]rule, len(apiMenuRules))
	for i, item := range apiMenuRules {
		sorted[i] = rule{item.prefix, item.menuPath}
	}
	sort.Slice(sorted, func(i, j int) bool { return len(sorted[i].prefix) > len(sorted[j].prefix) })
	for _, item := range sorted {
		if strings.HasPrefix(path, item.prefix) {
			return item.menuPath
		}
	}
	return ""
}

func deriveMenuID(path string, menuIDByPath map[string]int64) int64 {
	return menuIDByPath[deriveMenuPath(path)]
}

func syncSuperAdminRoleAPIs(db *gorm.DB) {
	var role po.RolePO
	db.Where("code = ?", "super-admin").First(&role)
	if role.ID == 0 {
		return
	}
	var existingIDs []int64
	db.Model(&po.RoleAPIPO{}).Where("role_id = ?", role.ID).Pluck("api_id", &existingIDs)
	existingSet := make(map[int64]bool, len(existingIDs))
	for _, id := range existingIDs {
		existingSet[id] = true
	}
	var allAPIs []po.APIPO
	db.Where("ignored = 0").Where("permission_mode = ? OR permission_mode = '' OR permission_mode IS NULL", "casbin").Find(&allAPIs)
	var toInsert []po.RoleAPIPO
	for _, api := range allAPIs {
		if !existingSet[api.ID] {
			toInsert = append(toInsert, po.RoleAPIPO{RoleID: role.ID, ApiID: api.ID})
		}
	}
	if len(toInsert) > 0 {
		db.Create(&toInsert)
	}
}

func deriveDescription(path, method string) string {
	return apiDescriptions[apiDescriptionKey{method: method, path: path}]
}

func deriveGroup(path string) string {
	type rule struct{ prefix, group string }
	sorted := make([]rule, len(groupRules))
	for i, item := range groupRules {
		sorted[i] = rule{item.prefix, item.group}
	}
	sort.Slice(sorted, func(i, j int) bool { return len(sorted[i].prefix) > len(sorted[j].prefix) })
	for _, item := range sorted {
		if strings.HasPrefix(path, item.prefix) {
			return item.group
		}
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}
