package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	BaseURL               = "https://api.dub.co"
	DefaultHTTPTimeout    = 30 * time.Second
	MaxRateLimitRetries   = 3
	RateLimitBaseDelay    = 1 * time.Second
	Max5xxRetries         = 1
	ServerErrorRetryDelay = 1 * time.Second

	// Circuit breaker constants
	CircuitBreakerThreshold = 5                // Open after 5 consecutive 5xx errors
	CircuitBreakerCooldown  = 30 * time.Second // Stay open for 30 seconds
)

// CircuitState represents the current state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Rejecting requests
	CircuitHalfOpen                     // Testing with a single request
)

// ErrCircuitOpen is returned when the circuit breaker is open and rejecting requests.
var ErrCircuitOpen = errors.New("circuit breaker is open: API server is experiencing issues")

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// Circuit breaker state
	cbMu               sync.RWMutex
	cbState            CircuitState
	cbConsecutive5xx   int
	cbOpenedAt         time.Time
	cbCooldown         time.Duration
	cbThreshold        int
	cbHalfOpenInFlight bool
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: BaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
			Transport: &http.Transport{
				MaxIdleConns:    100,
				MaxConnsPerHost: 10,
				IdleConnTimeout: 90 * time.Second,
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
		cbState:     CircuitClosed,
		cbCooldown:  CircuitBreakerCooldown,
		cbThreshold: CircuitBreakerThreshold,
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return c.doWithRetry(ctx, req)
}

func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	retries429 := 0
	retries5xx := 0
	isIdempotent := req.Method == "GET" || req.Method == "HEAD" || req.Method == "OPTIONS"

	// Generate a unique request ID for log correlation
	reqID := generateRequestID()

	for {
		// Check circuit breaker before making request
		if err := c.checkCircuitBreaker(); err != nil {
			return nil, err
		}

		slog.Debug("api request", "req_id", reqID, "method", req.Method, "url", req.URL.String())

		resp, err = c.httpClient.Do(req)
		if err != nil {
			slog.Debug("api request failed", "req_id", reqID, "error", err)
			return nil, err
		}

		slog.Debug("api response", "req_id", reqID, "status", resp.StatusCode)

		// 2xx: success, reset circuit breaker
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			c.recordSuccess()
			return resp, nil
		}

		// 4xx (except 429): no retry, but reset consecutive 5xx counter
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			c.recordSuccess() // 4xx is not a server error, reset counter
			return resp, nil
		}

		// 429: exponential backoff
		if resp.StatusCode == 429 {
			if retries429 >= MaxRateLimitRetries {
				return resp, nil
			}

			baseDelay := RateLimitBaseDelay * time.Duration(1<<retries429)
			jitter := time.Duration(mathrand.Int63n(int64(baseDelay / 2)))
			delay := baseDelay + jitter

			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}

			slog.Info("rate limited, retrying", "req_id", reqID, "delay", delay, "attempt", retries429+1)
			closeBody(resp)

			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to replay request body: %w", err)
				}
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			retries429++
			continue
		}

		// 5xx: record error and retry once for idempotent
		if resp.StatusCode >= 500 {
			c.record5xxError()

			if !isIdempotent || retries5xx >= Max5xxRetries {
				return resp, nil
			}

			slog.Info("retrying after server error", "req_id", reqID, "status", resp.StatusCode)
			closeBody(resp)

			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to replay request body: %w", err)
				}
			}

			select {
			case <-time.After(ServerErrorRetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			retries5xx++
			continue
		}

		return resp, nil
	}
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var getBody func() (io.ReadCloser, error)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.GetBody = getBody
	return c.Do(ctx, req)
}

func (c *Client) Patch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var getBody func() (io.ReadCloser, error)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, "PATCH", c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.GetBody = getBody
	return c.Do(ctx, req)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var getBody func() (io.ReadCloser, error)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.GetBody = getBody
	return c.Do(ctx, req)
}

func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

func (c *Client) DeleteWithBody(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var getBody func() (io.ReadCloser, error)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.GetBody = getBody
	return c.Do(ctx, req)
}

// APIKey returns the API key used by this client (for testing).
func (c *Client) APIKey() string {
	return c.apiKey
}

// Circuit breaker methods

// checkCircuitBreaker checks if a request should be allowed through.
// Returns nil if allowed, ErrCircuitOpen if the circuit is open and cooldown hasn't elapsed.
func (c *Client) checkCircuitBreaker() error {
	c.cbMu.Lock()
	defer c.cbMu.Unlock()

	switch c.cbState {
	case CircuitClosed:
		return nil
	case CircuitOpen:
		if time.Since(c.cbOpenedAt) >= c.cbCooldown {
			c.cbState = CircuitHalfOpen
			slog.Info("circuit breaker transitioning to half-open", "cooldown_elapsed", c.cbCooldown)
			return nil
		}
		remaining := c.cbCooldown - time.Since(c.cbOpenedAt)
		slog.Debug("circuit breaker is open", "remaining_cooldown", remaining)
		return ErrCircuitOpen
	case CircuitHalfOpen:
		if c.cbHalfOpenInFlight {
			return ErrCircuitOpen // Only one probe at a time
		}
		c.cbHalfOpenInFlight = true
		return nil
	}
	return nil
}

// recordSuccess records a successful request, resetting the circuit breaker to closed.
func (c *Client) recordSuccess() {
	c.cbMu.Lock()
	defer c.cbMu.Unlock()

	if c.cbState == CircuitHalfOpen {
		slog.Info("circuit breaker closing after successful half-open request")
	}
	c.cbState = CircuitClosed
	c.cbConsecutive5xx = 0
	c.cbHalfOpenInFlight = false
}

// record5xxError records a 5xx error and potentially opens the circuit breaker.
func (c *Client) record5xxError() {
	c.cbMu.Lock()
	defer c.cbMu.Unlock()

	c.cbConsecutive5xx++
	c.cbHalfOpenInFlight = false

	if c.cbState == CircuitHalfOpen {
		// Half-open test failed, reopen the circuit
		c.cbState = CircuitOpen
		c.cbOpenedAt = time.Now()
		slog.Warn("circuit breaker reopening after failed half-open request", "consecutive_5xx", c.cbConsecutive5xx)
		return
	}

	if c.cbConsecutive5xx >= c.cbThreshold {
		c.cbState = CircuitOpen
		c.cbOpenedAt = time.Now()
		slog.Warn("circuit breaker opening", "consecutive_5xx", c.cbConsecutive5xx, "threshold", c.cbThreshold)
	}
}

// CircuitState returns the current state of the circuit breaker (for testing).
func (c *Client) CircuitBreakerState() CircuitState {
	c.cbMu.RLock()
	defer c.cbMu.RUnlock()
	return c.cbState
}

// ResetCircuitBreaker resets the circuit breaker to closed state (for testing).
func (c *Client) ResetCircuitBreaker() {
	c.cbMu.Lock()
	defer c.cbMu.Unlock()
	c.cbState = CircuitClosed
	c.cbConsecutive5xx = 0
	c.cbOpenedAt = time.Time{}
	c.cbHalfOpenInFlight = false
}

func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// generateRequestID creates a short unique identifier for request correlation.
// Returns 8 hex characters (4 bytes of randomness).
func generateRequestID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
