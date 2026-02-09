package ratelimiter

import (
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
