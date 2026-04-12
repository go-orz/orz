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

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["name"] != "orz" {
		t.Fatalf("unexpected response body: %+v", response)
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

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["message"] != "bad request" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

func TestMessageUsesCustomMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := Message(ctx, http.StatusAccepted, "accepted"); err != nil {
		t.Fatalf("Message returned error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["message"] != "accepted" {
		t.Fatalf("unexpected response body: %+v", response)
	}
}

func TestMessageUsesStatusTextWhenMessageIsEmpty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := Message(ctx, http.StatusAccepted, ""); err != nil {
		t.Fatalf("Message returned error: %v", err)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["message"] != http.StatusText(http.StatusAccepted) {
		t.Fatalf("unexpected response body: %+v", response)
	}
}
