package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedactString_AccessKey(t *testing.T) {
	input := "aws_access_key_id=AKIAIOSFODNN7EXAMPLE"
	result := RedactString(input)
	if strings.Contains(result, "AKIAIOSFODNN7EXAMPLE") {
		t.Error("Access key should be redacted")
	}
	if !strings.Contains(result, "[REDACTED]") {
		t.Error("Expected [REDACTED] placeholder")
	}
}

func TestRedactString_SecretKey(t *testing.T) {
	input := "aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	result := RedactString(input)
	if strings.Contains(result, "wJalrXUtnFEMI") {
		t.Error("Secret key should be redacted")
	}
}

func TestRedactString_BearerToken(t *testing.T) {
	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	result := RedactString(input)
	if strings.Contains(result, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Error("Bearer token should be redacted")
	}
}

func TestRedactString_NoSensitiveData(t *testing.T) {
	input := "This is a normal log message with no secrets"
	result := RedactString(input)
	if result != input {
		t.Errorf("Expected unchanged string, got: %s", result)
	}
}

func TestRedactMap_SensitiveHeaders(t *testing.T) {
	input := map[string]string{
		"Authorization": "Bearer secret-token",
		"Content-Type":  "application/json",
		"X-Custom":      "some-value",
	}

	result := RedactMap(input)

	if result["Authorization"] != "[REDACTED]" {
		t.Errorf("Authorization should be redacted, got: %s", result["Authorization"])
	}
	if result["Content-Type"] != "application/json" {
		t.Errorf("Content-Type should not be redacted, got: %s", result["Content-Type"])
	}
}

func TestRedactHeaders(t *testing.T) {
	input := map[string][]string{
		"Authorization": {"Bearer secret"},
		"Content-Type":  {"application/json"},
	}

	result := RedactHeaders(input)

	if len(result["Authorization"]) != 1 || result["Authorization"][0] != "[REDACTED]" {
		t.Error("Authorization header should be redacted")
	}
	if len(result["Content-Type"]) != 1 || result["Content-Type"][0] != "application/json" {
		t.Error("Content-Type header should not be redacted")
	}
}

func TestIsSensitiveHeader(t *testing.T) {
	tests := []struct {
		header   string
		expected bool
	}{
		{"Authorization", true},
		{"authorization", true},
		{"X-API-Key", true},
		{"Cookie", true},
		{"Content-Type", false},
		{"Accept", false},
	}

	for _, tt := range tests {
		if IsSensitiveHeader(tt.header) != tt.expected {
			t.Errorf("IsSensitiveHeader(%q) = %v, want %v", tt.header, !tt.expected, tt.expected)
		}
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("Request ID should not be empty")
	}
	if id1 == id2 {
		t.Error("Request IDs should be unique")
	}
	if len(id1) != 8 {
		t.Errorf("Request ID length should be 8, got %d", len(id1))
	}
}

func TestGetRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, "test-123")
	if GetRequestID(ctx) != "test-123" {
		t.Errorf("Expected test-123, got %s", GetRequestID(ctx))
	}

	emptyCtx := context.Background()
	if GetRequestID(emptyCtx) != "" {
		t.Errorf("Expected empty string, got %s", GetRequestID(emptyCtx))
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Cache-Control":         "no-store",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("Header %s = %q, want %q", header, got, expected)
		}
	}
}

func TestRequestID(t *testing.T) {
	var capturedID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedID == "" {
		t.Error("Request ID should be set in context")
	}
	if w.Header().Get("X-Request-ID") != capturedID {
		t.Error("Response header should match context request ID")
	}
}

func TestRequestID_ExistingID(t *testing.T) {
	var capturedID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedID != "existing-123" {
		t.Errorf("Expected existing-123, got %s", capturedID)
	}
}

func TestRequireAPIKey_Missing(t *testing.T) {
	handler := RequireAPIKey("valid-key", nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestRequireAPIKey_Valid(t *testing.T) {
	handler := RequireAPIKey("valid-key", nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "valid-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestRequireAPIKey_QueryParam(t *testing.T) {
	handler := RequireAPIKey("valid-key", nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test?api_key=valid-key", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestRequireAPIKey_Disabled(t *testing.T) {
	handler := RequireAPIKey("", nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
