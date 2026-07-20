package middleware

import (
	"strings"
	"testing"
)

func TestSanitizeOperationLogBodyMasksSensitiveJSONFields(t *testing.T) {
	body := []byte(`{"username":"admin","password":"plain","profile":{"refresh_token":"secret-token","name":"Admin"}}`)

	got := sanitizeOperationLogBody(body)

	if strings.Contains(got, "plain") || strings.Contains(got, "secret-token") {
		t.Fatalf("sensitive values should be masked, got %s", got)
	}
	if !strings.Contains(got, `"password":"***"`) || !strings.Contains(got, `"refresh_token":"***"`) {
		t.Fatalf("sensitive keys should be kept with masked values, got %s", got)
	}
	if !strings.Contains(got, `"username":"admin"`) {
		t.Fatalf("non-sensitive values should be preserved, got %s", got)
	}
}

func TestSanitizeOperationLogBodyMasksSensitiveFormFields(t *testing.T) {
	body := []byte(`username=admin&newPassword=plain&access_token=secret-token`)

	got := sanitizeOperationLogBody(body)

	if strings.Contains(got, "plain") || strings.Contains(got, "secret-token") {
		t.Fatalf("sensitive values should be masked, got %s", got)
	}
	if !strings.Contains(got, "newPassword=%2A%2A%2A") || !strings.Contains(got, "access_token=%2A%2A%2A") {
		t.Fatalf("sensitive form fields should be masked, got %s", got)
	}
	if !strings.Contains(got, "username=admin") {
		t.Fatalf("non-sensitive values should be preserved, got %s", got)
	}
}
