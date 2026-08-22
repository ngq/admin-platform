package routesync

import (
	"net/http"
	"testing"

	gorp "github.com/ngq/gorp"
)

type stubRouter struct{}

func (stubRouter) Use(...gorp.Middleware)                                      {}
func (stubRouter) Group(string, ...gorp.Middleware) gorp.Router                { return stubRouter{} }
func (stubRouter) Handle(string, string, gorp.Handler, ...gorp.Middleware)     {}
func (stubRouter) HandleFunc(string, string, gorp.Handler, ...gorp.Middleware) {}
func (stubRouter) GET(string, gorp.Handler, ...gorp.Middleware)                {}
func (stubRouter) POST(string, gorp.Handler, ...gorp.Middleware)               {}
func (stubRouter) PUT(string, gorp.Handler, ...gorp.Middleware)                {}
func (stubRouter) DELETE(string, gorp.Handler, ...gorp.Middleware)             {}
func (stubRouter) PATCH(string, gorp.Handler, ...gorp.Middleware)              {}
func (stubRouter) HEAD(string, gorp.Handler, ...gorp.Middleware)               {}
func (stubRouter) OPTIONS(string, gorp.Handler, ...gorp.Middleware)            {}
func (stubRouter) ANY(string, gorp.Handler, ...gorp.Middleware)                {}
func (stubRouter) Mount(string, http.Handler)                                  {}
func (stubRouter) Static(string, string)                                       {}
func (stubRouter) StaticFile(string, string)                                   {}
func (stubRouter) StaticFS(string, http.FileSystem)                            {}
func (stubRouter) NoRoute(gorp.Handler)                                        {}
func (stubRouter) NoMethod(gorp.Handler)                                       {}
func (stubRouter) Routes() []gorp.RouteInfo                                    { return nil }

func TestRegistryTracksContractRoutes(t *testing.T) {
	registry := NewRegistry()
	router := registry.Wrap(stubRouter{})
	users := router.Group("/api/v1").Group("/users")
	users.GET("", func(gorp.Context) {})
	users.POST("/:id", func(gorp.Context) {})

	routes := registry.Routes()
	if len(routes) != 2 {
		t.Fatalf("got %d routes, want 2", len(routes))
	}
	if routes[0] != (Route{Method: http.MethodGet, Path: "/api/v1/users"}) {
		t.Fatalf("unexpected first route: %#v", routes[0])
	}
	if routes[1] != (Route{Method: http.MethodPost, Path: "/api/v1/users/:id"}) {
		t.Fatalf("unexpected second route: %#v", routes[1])
	}
}

func TestDeriveMenuPath(t *testing.T) {
	t.Parallel()
	tests := []struct{ path, want string }{
		{"/api/v1/roles/permission-config", "/system/role"},
		{"/api/v1/apis/batch-delete", "/system/api"},
		{"/api/v1/operation-logs/batch-delete", "/logs/operation"},
		{"/api/v1/auth/login", ""},
	}
	for _, test := range tests {
		if got := deriveMenuPath(test.path); got != test.want {
			t.Fatalf("deriveMenuPath(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}

func TestDeriveMenuID(t *testing.T) {
	t.Parallel()
	menus := map[string]int64{"/system/user": 311}
	if got := deriveMenuID("/api/v1/users", menus); got != 311 {
		t.Fatalf("deriveMenuID users = %d, want 311", got)
	}
	if got := deriveMenuID("/api/v1/auth/login", menus); got != 0 {
		t.Fatalf("deriveMenuID login = %d, want 0", got)
	}
}
