package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func TestLongCacheLifetime(t *testing.T) {
	e := echo.New()
	e.Use(longCacheLifetime)
	e.GET("/", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if got := rec.Header().Get(echo.HeaderCacheControl); got != "public, max-age=31536000" {
		t.Errorf("Cache-Control = %q, want %q", got, "public, max-age=31536000")
	}
}

// runWithIP drives the middleware directly with a given X-Real-IP and a handler
// that completes immediately, returning the response status.
func runWithIP(t *testing.T, mw echo.MiddlewareFunc, ip string) int {
	t.Helper()
	e := echo.NewWithConfig(echo.Config{
		IPExtractor: echo.ExtractIPFromRealIPHeader(),
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", ip)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mw(func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	return rec.Code
}

func TestSSEConnectionLimiter(t *testing.T) {
	mw := sseConnectionLimiter(1)

	// First request acquires the slot and blocks until released.
	release := make(chan struct{})
	done := make(chan int, 1)
	go func() {
		e := echo.NewWithConfig(echo.Config{IPExtractor: echo.ExtractIPFromRealIPHeader()})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handler := mw(func(c *echo.Context) error {
			<-release
			return c.NoContent(http.StatusOK)
		})
		_ = handler(c)
		done <- rec.Code
	}()

	// Ensure the first request has grabbed its slot.
	time.Sleep(50 * time.Millisecond)

	// A second request from the same IP is rejected while the first is active.
	if got := runWithIP(t, mw, "10.0.0.1"); got != http.StatusServiceUnavailable {
		t.Errorf("second request from same IP = %d, want %d", got, http.StatusServiceUnavailable)
	}

	// A different IP is not limited.
	if got := runWithIP(t, mw, "10.0.0.2"); got != http.StatusOK {
		t.Errorf("different IP status = %d, want %d", got, http.StatusOK)
	}

	// Release the first connection; it completes with 200.
	close(release)
	select {
	case code := <-done:
		if code != http.StatusOK {
			t.Errorf("first request status after release = %d, want %d", code, http.StatusOK)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first request did not complete after release")
	}

	// After release, a new request from the same IP succeeds.
	if got := runWithIP(t, mw, "10.0.0.1"); got != http.StatusOK {
		t.Errorf("request after release = %d, want %d", got, http.StatusOK)
	}
}
