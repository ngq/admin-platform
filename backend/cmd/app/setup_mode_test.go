package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	svcsetup "github.com/ngq/admin-platform/internal/service/setup"
	gorp "github.com/ngq/gorp"
)

func TestProbeDatabaseReadiness_DetectsMissingEnvAndPlaceholder(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}

	appYaml := `database:
  driver: mysql
  dsn: env(DATABASE_DSN)
`
	if err := os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte(appYaml), 0o644); err != nil {
		t.Fatalf("write app.yaml: %v", err)
	}

	// Unset any inherited DATABASE_DSN for isolated probe test
	oldDSN, hadDSN := os.LookupEnv("DATABASE_DSN")
	_ = os.Unsetenv("DATABASE_DSN")
	defer func() {
		if hadDSN {
			_ = os.Setenv("DATABASE_DSN", oldDSN)
		} else {
			_ = os.Unsetenv("DATABASE_DSN")
		}
	}()

	res := probeDatabaseReadiness(tmp)
	if !res.needsSetup {
		t.Fatalf("expected needsSetup=true when DATABASE_DSN is missing, got %#v", res)
	}

	// Placeholder .env should also trigger setup mode
	envContent := "DATABASE_DSN=user:password@tcp(127.0.0.1:3306)/admin_platform?charset=utf8mb4\n"
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	res = probeDatabaseReadiness(tmp)
	if !res.needsSetup {
		t.Fatalf("expected needsSetup=true for placeholder .env, got %#v", res)
	}
}

func TestExpandEnvPlaceholders(t *testing.T) {
	t.Setenv("TEST_CUSTOM_DSN", "root:123@tcp(127.0.0.1:3306)/app")
	got, unresolved := expandEnvPlaceholders("env(TEST_CUSTOM_DSN)")
	if unresolved || got != "root:123@tcp(127.0.0.1:3306)/app" {
		t.Fatalf("unexpected env() result: got=%q unresolved=%v", got, unresolved)
	}

	got, unresolved = expandEnvPlaceholders("${MISSING_VAR:-fallback}")
	if unresolved || got != "fallback" {
		t.Fatalf("unexpected default ${VAR:-default} result: got=%q unresolved=%v", got, unresolved)
	}

	_, unresolved = expandEnvPlaceholders("env(DEFINITELY_MISSING_VAR_XYZ)")
	if !unresolved {
		t.Fatal("expected unresolved=true for missing env() var")
	}
}

func TestSetupModeBuildHTTPRuntime_WithEmptyYAMLDSNAndNoRoute(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}

	appYAML := `app:
  name: admin-platform
  env: dev
server:
  http:
    addr: 127.0.0.1:0
    mode: debug
database:
  driver: mysql
  dsn: ""
`
	if err := os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte(appYAML), 0o644); err != nil {
		t.Fatalf("write app.yaml: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()

	cleanup := ensureSetupModePlaceholderEnv()
	defer cleanup()

	setupSvc := svcsetup.NewService(svcsetup.Options{
		Initialized: false,
		Reason:      "database not configured",
	})
	rt, err := gorp.BuildHTTPRuntime(setupWizardRuntimeOptions(setupSvc)...)
	if err != nil {
		t.Fatalf("BuildHTTPRuntime in setup mode failed: %v", err)
	}
	if rt.Container != nil {
		defer func() { _ = rt.Container.Destroy() }()
	}
	if err := registerSetupWizardRoutes(rt, setupSvc); err != nil {
		t.Fatalf("registerSetupWizardRoutes failed: %v", err)
	}

	httpSvc := rt.HTTP()
	if httpSvc == nil || httpSvc.Server() == nil || httpSvc.Server().Handler == nil {
		t.Fatal("expected initialized HTTP server handler in setup mode")
	}

	// Setup status endpoint must work in setup mode.
	statusRec := httptest.NewRecorder()
	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	httpSvc.Server().Handler.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/setup/status, got %d body=%s", statusRec.Code, statusRec.Body.String())
	}

	// Non-setup endpoints must be blocked by NoRoute with 503 + SETUP_REQUIRED.
	loginRec := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	httpSvc.Server().Handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for blocked route in setup mode, got %d body=%s", loginRec.Code, loginRec.Body.String())
	}
	var errResp struct {
		Code   int    `json:"code"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal 503 response: %v", err)
	}
	if errResp.Reason != "SETUP_REQUIRED" {
		t.Fatalf("expected reason SETUP_REQUIRED, got %+v", errResp)
	}
}
