package main

import (
	"net/http"
	"time"

	"github.com/go-rate-limiter/internal/ratelimiter"
)

func index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Request Allowed!"))
}

func main() {
	rateLimiter := ratelimiter.New()
	rateLimiter.ConfigureClient("client-1", 5, time.Minute)
	rateLimiter.ConfigureClient("client-2", 5, time.Minute)

	http.HandleFunc("/", rateLimiter.Limit(index))
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}