package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	jwtprovider "github.com/ngq/gorp/framework/provider/auth/jwt"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
)

const casbinTestModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

func TestCasbinRBAC_AbortsWhenClaimsMissing(t *testing.T) {
	engine, called := newCasbinTestEngine(t, nil, false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	engine.ServeHTTP(rec, req)

	if *called {
		t.Fatal("handler should not run when claims are missing")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("response should not include handler output: %s", rec.Body.String())
	}
}

func TestCasbinRBAC_AbortsWhenPolicyMissing(t *testing.T) {
	claims := &securitycontract.JWTClaims{
		SubjectID:   1,
		SubjectType: "user",
		SubjectName: "limited",
		Roles:       []string{"limited-ai"},
	}
	engine, called := newCasbinTestEngine(t, claims, false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	engine.ServeHTTP(rec, req)

	if *called {
		t.Fatal("handler should not run when Casbin denies access")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("response should not include handler output: %s", rec.Body.String())
	}
}

func TestCasbinRBAC_AllowsWhenPolicyMatches(t *testing.T) {
	claims := &securitycontract.JWTClaims{
		SubjectID:   1,
		SubjectType: "user",
		SubjectName: "limited",
		Roles:       []string{"limited-ai"},
	}
	engine, called := newCasbinTestEngine(t, claims, true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	engine.ServeHTTP(rec, req)

	if !*called {
		t.Fatal("handler should run when Casbin allows access")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthMiddleware_AbortsWhenTokenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := ginprovider.NewTestEngine()
	jwtSvc := jwtprovider.NewJWTService("secret", "issuer", "")
	called := false
	engine.Use(ginprovider.AdaptMiddleware(AuthMiddleware(jwtSvc, "user", nil)))
	engine.GET("/api/v1/users", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	engine.ServeHTTP(rec, req)

	if called {
		t.Fatal("handler should not run when bearer token is missing")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("response should not include handler output: %s", rec.Body.String())
	}
}

func newCasbinTestEngine(t *testing.T, claims *securitycontract.JWTClaims, allow bool) (*gin.Engine, *bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	m, err := model.NewModelFromString(casbinTestModel)
	if err != nil {
		t.Fatalf("create casbin model: %v", err)
	}
	enforcer, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		t.Fatalf("create casbin enforcer: %v", err)
	}
	if allow && claims != nil && len(claims.Roles) > 0 {
		_, _ = enforcer.AddPolicy(claims.Roles[0], "/api/v1/users", http.MethodGet)
	}

	called := false
	engine := ginprovider.NewTestEngine()
	if claims != nil {
		engine.Use(func(c *gin.Context) {
			c.Set(securitycontract.ContextJWTClaimsKey, claims)
			c.Next()
		})
	}
	engine.Use(ginprovider.AdaptMiddleware(CasbinRBAC(enforcer)))
	engine.GET("/api/v1/users", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return engine, &called
}
