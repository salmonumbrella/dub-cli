package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("dub_test123")
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer dub_test123" {
			t.Errorf("expected Bearer auth header")
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL

	resp, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// Circuit Breaker Tests

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL
	client.cbThreshold = 5
	client.cbCooldown = 1 * time.Second

	ctx := context.Background()

	// Make requests until circuit opens
	// With retry logic, each request makes 2 actual HTTP calls (original + 1 retry for GET)
	// So we need (threshold) / 2 = 3 requests to hit 5 consecutive 5xx
	// Actually, we track consecutive 5xx at the response level, so with Max5xxRetries=1,
	// each failed request records 2 errors (original + retry), so 3 requests = 6 errors >= 5
	for i := 0; i < 3; i++ {
		resp, err := client.Get(ctx, "/test")
		if err != nil && !errors.Is(err, ErrCircuitOpen) {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	// Circuit should now be open
	if client.CircuitBreakerState() != CircuitOpen {
		t.Errorf("expected circuit to be open, got %v", client.CircuitBreakerState())
	}

	// Next request should fail immediately with ErrCircuitOpen
	_, err := client.Get(ctx, "/test")
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 4 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL
	client.cbThreshold = 5
	client.cbCooldown = 100 * time.Millisecond

	ctx := context.Background()

	// Make 2 requests that fail (each makes 2 HTTP calls due to retry = 4 5xx errors)
	for i := 0; i < 2; i++ {
		resp, _ := client.Get(ctx, "/test")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	// Circuit should still be closed (4 errors < 5 threshold)
	if client.CircuitBreakerState() != CircuitClosed {
		t.Errorf("expected circuit to be closed, got %v", client.CircuitBreakerState())
	}

	// Make a successful request
	resp, err := client.Get(ctx, "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	// Circuit should be closed and counter reset
	if client.CircuitBreakerState() != CircuitClosed {
		t.Errorf("expected circuit to be closed after success, got %v", client.CircuitBreakerState())
	}
}

func TestCircuitBreaker_HalfOpenTransition(t *testing.T) {
	var requestCount int32
	var shouldSucceed int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		if atomic.LoadInt32(&shouldSucceed) == 1 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL
	client.cbThreshold = 5
	client.cbCooldown = 100 * time.Millisecond

	ctx := context.Background()

	// Make enough requests to open the circuit
	for i := 0; i < 3; i++ {
		resp, _ := client.Get(ctx, "/test")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	if client.CircuitBreakerState() != CircuitOpen {
		t.Fatalf("expected circuit to be open, got %v", client.CircuitBreakerState())
	}

	// Wait for cooldown
	time.Sleep(150 * time.Millisecond)

	// Set server to succeed
	atomic.StoreInt32(&shouldSucceed, 1)

	// Make request - should go through (half-open) and succeed
	resp, err := client.Get(ctx, "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp.Body.Close()

	// Circuit should now be closed
	if client.CircuitBreakerState() != CircuitClosed {
		t.Errorf("expected circuit to be closed after half-open success, got %v", client.CircuitBreakerState())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL
	client.cbThreshold = 5
	client.cbCooldown = 100 * time.Millisecond

	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 3; i++ {
		resp, _ := client.Get(ctx, "/test")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	if client.CircuitBreakerState() != CircuitOpen {
		t.Fatalf("expected circuit to be open, got %v", client.CircuitBreakerState())
	}

	// Wait for cooldown
	time.Sleep(150 * time.Millisecond)

	// Make request - should go through (half-open) but fail
	resp, err := client.Get(ctx, "/test")
	if err != nil && !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	// Circuit should be open again
	if client.CircuitBreakerState() != CircuitOpen {
		t.Errorf("expected circuit to be open after half-open failure, got %v", client.CircuitBreakerState())
	}
}

func TestCircuitBreaker_4xxDoesNotTrigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL
	client.cbThreshold = 2 // Low threshold for quick test

	ctx := context.Background()

	// Make multiple 4xx requests
	for i := 0; i < 10; i++ {
		resp, err := client.Get(ctx, "/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = resp.Body.Close()
	}

	// Circuit should still be closed
	if client.CircuitBreakerState() != CircuitClosed {
		t.Errorf("expected circuit to remain closed for 4xx errors, got %v", client.CircuitBreakerState())
	}
}

func TestCircuitBreaker_RejectsRequestsWhenOpen(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL
	client.cbThreshold = 5
	client.cbCooldown = 10 * time.Second // Long cooldown

	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 3; i++ {
		resp, _ := client.Get(ctx, "/test")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	initialCount := atomic.LoadInt32(&requestCount)

	// Try to make more requests - should be rejected immediately
	for i := 0; i < 5; i++ {
		_, err := client.Get(ctx, "/test")
		if !errors.Is(err, ErrCircuitOpen) {
			t.Errorf("expected ErrCircuitOpen, got %v", err)
		}
	}

	// No additional HTTP requests should have been made
	if atomic.LoadInt32(&requestCount) != initialCount {
		t.Errorf("expected no additional requests, got %d more", atomic.LoadInt32(&requestCount)-initialCount)
	}
}

func TestCircuitBreaker_ResetCircuitBreaker(t *testing.T) {
	client := NewClient("dub_test123")
	client.cbState = CircuitOpen
	client.cbConsecutive5xx = 10
	client.cbOpenedAt = time.Now()

	client.ResetCircuitBreaker()

	if client.CircuitBreakerState() != CircuitClosed {
		t.Errorf("expected circuit to be closed after reset, got %v", client.CircuitBreakerState())
	}
}
