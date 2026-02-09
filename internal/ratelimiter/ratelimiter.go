package ratelimiter

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ClientWindow struct {
	WindowStart time.Time
	Count       int
	Limit       int
	Window      time.Duration
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*ClientWindow
}

func New() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientWindow),
	}
}

func (rl *RateLimiter) ConfigureClient(
	clientID string,
	limit int,
	window time.Duration,
) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.clients[clientID] = &ClientWindow{
		WindowStart: time.Now(),
		Count:       0,
		Limit:       limit,
		Window:      window,
	}
}

func (rl *RateLimiter) Allow(clientID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.clients[clientID]
	if !exists {
		return false, fmt.Errorf("client not configured")
	}

	now := time.Now()

	if now.Sub(client.WindowStart) >= client.Window {
		client.WindowStart = now
		client.Count = 0
	}

	if client.Count >= client.Limit {
		return false, nil
	}

	client.Count++
	return true, nil
}

func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.Header.Get("X-Client-ID")
		if clientID == "" {
			http.Error(w, "Missing X-Client-ID", http.StatusBadRequest)
			return
		}

		allowed, err := rl.Allow(clientID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	})
}