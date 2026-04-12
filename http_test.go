package orz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestOkUsesResponseEnvelope(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := Ok(ctx, map[string]string{"name": "orz"}); err != nil {
		t.Fatalf("Ok returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Code != http.StatusOK || response.Message != "ok" {
		t.Fatalf("unexpected response envelope: %+v", response)
	}
}

func TestErrorResponseUsesHTTPStatus(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := ErrorResponse(ctx, http.StatusBadRequest, "bad request"); err != nil {
		t.Fatalf("ErrorResponse returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var response Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Code != http.StatusBadRequest || response.Message != "bad request" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestRespondUsesCustomMessageAndData(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := Respond(ctx, http.StatusAccepted, "accepted", map[string]string{"task": "sync"}); err != nil {
		t.Fatalf("Respond returned error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}

	var response Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Code != http.StatusAccepted || response.Message != "accepted" {
		t.Fatalf("unexpected response envelope: %+v", response)
	}
}
