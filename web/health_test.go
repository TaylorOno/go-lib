package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_JSONResponse(t *testing.T) {
	s := NewServer()
	s.RegisterReadinessCheck("test-check", func(ctx context.Context) error {
		return errors.New("failed")
	})

	req := httptest.NewRequest("GET", "/readyz", nil)
	recorder := httptest.NewRecorder()

	handler := s.readinessHandler()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if response["status"] != "DOWN" {
		t.Errorf("expected status DOWN, got %v", response["status"])
	}

	checks, ok := response["checks"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected checks to be a map, got %T", response["checks"])
	}

	if checks["test-check"] != "failed" {
		t.Errorf("expected test-check to be 'failed', got %v", checks["test-check"])
	}
}

func TestLivenessHandler_SimpleOK(t *testing.T) {
	s := NewServer()
	// No liveness checks registered by default

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	handler := s.livenessHandler()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
