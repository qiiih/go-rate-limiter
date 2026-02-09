package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllowWithinLimit(t *testing.T) {
	rl := New()
	rl.ConfigureClient("test", 2, time.Minute)

	ok, _ := rl.Allow("test")
	if !ok {
		t.Fatal("expected request to be allowed")
	}

	ok, _ = rl.Allow("test")
	if !ok {
		t.Fatal("expected request to be allowed")
	}
}

func TestRejectAfterLimit(t *testing.T) {
	rl := New()
	rl.ConfigureClient("test", 1, time.Minute)

	ok, _ := rl.Allow("test")
	if !ok {
		t.Fatal("expected first request allowed")
	}

	ok, _ = rl.Allow("test")
	if ok {
		t.Fatal("expected request rejected")
	}
}

func TestWindowReset(t *testing.T) {
	rl := New()
	rl.ConfigureClient("test", 1, 50*time.Millisecond)

	rl.Allow("test")
	time.Sleep(60 * time.Millisecond)

	ok, _ := rl.Allow("test")
	if !ok {
		t.Fatal("expected request allowed after window reset")
	}
}

func TestUnknownClient(t *testing.T) {
	rl := New()

	_, err := rl.Allow("unknown")
	if err == nil {
		t.Fatal("expected error for unknown client")
	}
}

func TestMiddlewareMissingClientID(t *testing.T) {
	rl := New()

	handler := rl.Limit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestMiddlewareAllowsRequest(t *testing.T) {
	rl := New()
	rl.ConfigureClient("client1", 1, time.Minute)

	handler := rl.Limit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-ID", "client1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestMiddlewareBlocksAfterLimit(t *testing.T) {
	rl := New()
	rl.ConfigureClient("client1", 1, time.Minute)

	handler := rl.Limit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Client-ID", "client1")

	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec2.Code)
	}
}


