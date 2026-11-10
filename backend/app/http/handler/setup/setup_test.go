package setup

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	svcsetup "github.com/ngq/admin-platform/internal/service/setup"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
)

func newSetupTestEngine(t *testing.T, svc *svcsetup.Service) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := ginprovider.NewTestEngine()
	h := NewHandler(svc)
	v1 := engine.Group("/api/v1/setup")
	v1.GET("/status", ginprovider.AdaptHandler(h.Status))
	v1.POST("/test-connection", ginprovider.AdaptHandler(h.TestConnection))
	v1.POST("/database", ginprovider.AdaptHandler(h.InitializeDatabase))
	return engine
}

func TestSetupStatusHandler_NoHardcodedDefaults(t *testing.T) {
	svc := svcsetup.NewService(svcsetup.Options{
		Initialized: false,
		Reason:      "database not configured",
	})
	engine := newSetupTestEngine(t, svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 0 || resp.Data["initialized"] != false {
		t.Fatalf("unexpected setup status payload: %+v", resp)
	}
	for _, forbiddenKey := range []string{"default_host", "default_port", "default_database"} {
		if _, exists := resp.Data[forbiddenKey]; exists {
			t.Fatalf("status payload should not hardcode %s", forbiddenKey)
		}
	}
}

func TestSetupHandlers_AlwaysRejectWhenInitialized(t *testing.T) {
	svc := svcsetup.NewService(svcsetup.Options{
		Initialized:  true,
		ActiveDriver: "mysql",
		ActiveDSN:    "root:secret@tcp(127.0.0.1:3306)/admin_platform",
	})
	engine := newSetupTestEngine(t, svc)

	payload := []byte(`{"driver":"mysql","host":"127.0.0.1","port":3306,"username":"root","password":"pwd","database":"admin_platform"}`)
	for _, path := range []string{"/api/v1/setup/test-connection", "/api/v1/setup/database"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s expected 403 when already initialized, got %d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}
