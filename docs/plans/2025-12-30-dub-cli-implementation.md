# Dub CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an agent-friendly Go CLI that mirrors the full Dub API (39 endpoints) with keyring-based auth.

**Architecture:** Cobra CLI with modular command structure. API client handles auth, retries, rate limiting. Credentials stored in OS keyring. Browser-based setup UI for interactive auth.

**Tech Stack:** Go 1.24, Cobra, 99designs/keyring, gojq, termenv, GoReleaser

---

## Task 1: Project Scaffolding

**Files:**
- Create: `cmd/dub/main.go`
- Create: `internal/config/paths.go`
- Create: `Makefile`

**Step 1: Create main.go entry point**

```go
// cmd/dub/main.go
package main

import (
	"os"

	"github.com/salmonumbrella/dub-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
```

**Step 2: Create config paths**

```go
// internal/config/paths.go
package config

const (
	AppName = "dub-cli"
)
```

**Step 3: Create Makefile**

```makefile
.PHONY: build test lint fmt

build:
	go build -o dub ./cmd/dub

test:
	go test -v ./...

lint:
	golangci-lint run

fmt:
	gofumpt -w .
	goimports -w .
```

**Step 4: Verify build compiles (will fail - cmd package missing)**

Run: `go build ./cmd/dub`
Expected: FAIL with "package cmd not found"

**Step 5: Commit scaffolding**

```bash
git add cmd/dub/main.go internal/config/paths.go Makefile
git commit -m "feat: add project scaffolding"
```

---

## Task 2: Root Command with Global Flags

**Files:**
- Create: `internal/cmd/root.go`
- Create: `internal/cmd/root_test.go`

**Step 1: Write failing test for root command**

```go
// internal/cmd/root_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("dub")) {
		t.Errorf("expected output to contain 'dub', got: %s", output)
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	cmd := NewRootCmd()

	// Check persistent flags exist
	flags := []string{"workspace", "output", "query", "yes", "debug", "limit", "sort-by", "desc"}
	for _, name := range flags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected persistent flag %q to exist", name)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/... -v -run TestRootCommand`
Expected: FAIL with "NewRootCmd not defined"

**Step 3: Implement root command**

```go
// internal/cmd/root.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type rootFlags struct {
	Workspace string
	Output    string
	Query     string
	Yes       bool
	Debug     bool
	Limit     int
	SortBy    string
	Desc      bool
	Color     string
}

var flags rootFlags

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dub",
		Short:        "Dub CLI - manage your Dub links from the terminal",
		Long:         "A command-line interface for the Dub API. Manage links, analytics, domains, and more.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.Desc && flags.SortBy == "" {
				return fmt.Errorf("--desc requires --sort-by to be specified")
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&flags.Workspace, "workspace", "w", os.Getenv("DUB_WORKSPACE"), "Workspace name (or DUB_WORKSPACE env)")
	cmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", getEnvOrDefault("DUB_OUTPUT", "text"), "Output format: text|json")
	cmd.PersistentFlags().StringVar(&flags.Query, "query", "", "JQ filter expression for JSON output")
	cmd.PersistentFlags().BoolVarP(&flags.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.PersistentFlags().BoolVar(&flags.Yes, "force", false, "Skip confirmation prompts (alias for --yes)")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug output")
	cmd.PersistentFlags().IntVar(&flags.Limit, "limit", 0, "Limit number of results (0 = no limit)")
	cmd.PersistentFlags().StringVar(&flags.SortBy, "sort-by", "", "Field name to sort by")
	cmd.PersistentFlags().BoolVar(&flags.Desc, "desc", false, "Sort descending (requires --sort-by)")
	cmd.PersistentFlags().StringVar(&flags.Color, "color", "auto", "Color output: auto|always|never")

	return cmd
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Execute(args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func ExecuteContext(ctx context.Context, args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}
```

**Step 4: Get dependencies and run tests**

Run: `go mod tidy && go test ./internal/cmd/... -v -run TestRootCommand`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cmd/root.go internal/cmd/root_test.go go.mod go.sum
git commit -m "feat: add root command with global flags"
```

---

## Task 3: Secrets Store (Keyring)

**Files:**
- Create: `internal/secrets/store.go`
- Create: `internal/secrets/store_test.go`

**Step 1: Write failing test**

```go
// internal/secrets/store_test.go
package secrets

import (
	"testing"
	"time"
)

func TestCredentials_Fields(t *testing.T) {
	creds := Credentials{
		Name:      "test-workspace",
		APIKey:    "dub_test123",
		CreatedAt: time.Now(),
	}

	if creds.Name != "test-workspace" {
		t.Errorf("expected name 'test-workspace', got %q", creds.Name)
	}
	if creds.APIKey != "dub_test123" {
		t.Errorf("expected api key 'dub_test123', got %q", creds.APIKey)
	}
}

func TestParseCredentialKey(t *testing.T) {
	tests := []struct {
		key      string
		wantName string
		wantOK   bool
	}{
		{"workspace:production", "production", true},
		{"workspace:dev", "dev", true},
		{"other:key", "", false},
		{"workspace:", "", false},
	}

	for _, tt := range tests {
		name, ok := ParseCredentialKey(tt.key)
		if ok != tt.wantOK || name != tt.wantName {
			t.Errorf("ParseCredentialKey(%q) = (%q, %v), want (%q, %v)",
				tt.key, name, ok, tt.wantName, tt.wantOK)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/secrets/... -v`
Expected: FAIL with "Credentials not defined"

**Step 3: Implement secrets store**

```go
// internal/secrets/store.go
package secrets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/dub-cli/internal/config"
)

type Store interface {
	Keys() ([]string, error)
	Set(name string, creds Credentials) error
	Get(name string) (Credentials, error)
	Delete(name string) error
	List() ([]Credentials, error)
}

type KeyringStore struct {
	ring keyring.Keyring
}

type Credentials struct {
	Name      string    `json:"name"`
	APIKey    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type storedCredentials struct {
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func OpenDefault() (Store, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: config.AppName,
	})
	if err != nil {
		return nil, err
	}
	return &KeyringStore{ring: ring}, nil
}

func (s *KeyringStore) Keys() ([]string, error) {
	return s.ring.Keys()
}

func (s *KeyringStore) Set(name string, creds Credentials) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing workspace name")
	}
	if creds.APIKey == "" {
		return fmt.Errorf("missing API key")
	}
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(storedCredentials{
		APIKey:    creds.APIKey,
		CreatedAt: creds.CreatedAt,
	})
	if err != nil {
		return err
	}

	return s.ring.Set(keyring.Item{
		Key:  credentialKey(name),
		Data: payload,
	})
}

func (s *KeyringStore) Get(name string) (Credentials, error) {
	name = normalize(name)
	if name == "" {
		return Credentials{}, fmt.Errorf("missing workspace name")
	}
	item, err := s.ring.Get(credentialKey(name))
	if err != nil {
		return Credentials{}, err
	}
	var stored storedCredentials
	if err := json.Unmarshal(item.Data, &stored); err != nil {
		return Credentials{}, err
	}

	return Credentials{
		Name:      name,
		APIKey:    stored.APIKey,
		CreatedAt: stored.CreatedAt,
	}, nil
}

func (s *KeyringStore) Delete(name string) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing workspace name")
	}
	return s.ring.Remove(credentialKey(name))
}

func (s *KeyringStore) List() ([]Credentials, error) {
	keys, err := s.Keys()
	if err != nil {
		return nil, err
	}
	var out []Credentials
	for _, k := range keys {
		name, ok := ParseCredentialKey(k)
		if !ok {
			continue
		}
		creds, err := s.Get(name)
		if err != nil {
			return nil, err
		}
		out = append(out, creds)
	}
	return out, nil
}

func ParseCredentialKey(k string) (name string, ok bool) {
	const prefix = "workspace:"
	if !strings.HasPrefix(k, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(k, prefix)
	if strings.TrimSpace(rest) == "" {
		return "", false
	}
	return rest, true
}

func credentialKey(name string) string {
	return fmt.Sprintf("workspace:%s", name)
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
```

**Step 4: Get dependencies and run tests**

Run: `go mod tidy && go test ./internal/secrets/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/secrets/store.go internal/secrets/store_test.go go.mod go.sum
git commit -m "feat: add keyring-based secrets store"
```

---

## Task 4: API Client Core

**Files:**
- Create: `internal/api/client.go`
- Create: `internal/api/client_test.go`
- Create: `internal/api/errors.go`

**Step 1: Write failing test**

```go
// internal/api/client_test.go
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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
		w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	client := NewClient("dub_test123")
	client.baseURL = server.URL

	resp, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/api/... -v`
Expected: FAIL with "NewClient not defined"

**Step 3: Implement API client**

```go
// internal/api/client.go
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	BaseURL               = "https://api.dub.co"
	DefaultHTTPTimeout    = 30 * time.Second
	MaxRateLimitRetries   = 3
	RateLimitBaseDelay    = 1 * time.Second
	Max5xxRetries         = 1
	ServerErrorRetryDelay = 1 * time.Second
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
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

	for {
		slog.Debug("api request", "method", req.Method, "url", req.URL.String())

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// 4xx (except 429): no retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
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

			slog.Info("rate limited, retrying", "delay", delay, "attempt", retries429+1)
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

		// 5xx: retry once for idempotent
		if resp.StatusCode >= 500 {
			if !isIdempotent || retries5xx >= Max5xxRetries {
				return resp, nil
			}

			slog.Info("retrying after server error", "status", resp.StatusCode)
			closeBody(resp)

			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to replay request body: %w", err)
				}
			}

			time.Sleep(ServerErrorRetryDelay)
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

func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}
```

**Step 4: Create errors.go**

```go
// internal/api/errors.go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	DocURL  string `json:"doc_url"`
}

type errorResponse struct {
	Error APIError `json:"error"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Message
}

func ParseAPIError(body []byte) *APIError {
	var resp errorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return &APIError{Message: string(body)}
	}
	return &resp.Error
}

func ReadAPIError(resp *http.Response) *APIError {
	body, _ := io.ReadAll(resp.Body)
	return ParseAPIError(body)
}

func WrapError(method, url string, status int, err error) error {
	return fmt.Errorf("%s %s returned %d: %w", method, url, status, err)
}
```

**Step 5: Run tests**

Run: `go mod tidy && go test ./internal/api/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/api/client.go internal/api/client_test.go internal/api/errors.go go.mod go.sum
git commit -m "feat: add API client with retry logic"
```

---

## Task 5: Output Formatting

**Files:**
- Create: `internal/outfmt/format.go`
- Create: `internal/outfmt/format_test.go`

**Step 1: Write failing test**

```go
// internal/outfmt/format_test.go
package outfmt

import (
	"bytes"
	"testing"
)

func TestFormatJSON(t *testing.T) {
	data := map[string]string{"id": "123", "url": "https://dub.sh/test"}
	buf := new(bytes.Buffer)

	err := FormatJSON(buf, data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte(`"id"`)) {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

func TestFormatJSON_WithQuery(t *testing.T) {
	data := map[string]string{"id": "123", "url": "https://dub.sh/test"}
	buf := new(bytes.Buffer)

	err := FormatJSON(buf, data, ".id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "\"123\"\n" {
		t.Errorf("expected '\"123\"\\n', got: %q", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/outfmt/... -v`
Expected: FAIL with "FormatJSON not defined"

**Step 3: Implement output formatting**

```go
// internal/outfmt/format.go
package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

type contextKey string

const (
	formatKey  contextKey = "format"
	queryKey   contextKey = "query"
	yesKey     contextKey = "yes"
	limitKey   contextKey = "limit"
	sortByKey  contextKey = "sortBy"
	descKey    contextKey = "desc"
)

func WithFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, formatKey, format)
}

func GetFormat(ctx context.Context) string {
	if v, ok := ctx.Value(formatKey).(string); ok {
		return v
	}
	return "text"
}

func WithQuery(ctx context.Context, query string) context.Context {
	return context.WithValue(ctx, queryKey, query)
}

func GetQuery(ctx context.Context) string {
	if v, ok := ctx.Value(queryKey).(string); ok {
		return v
	}
	return ""
}

func WithYes(ctx context.Context, yes bool) context.Context {
	return context.WithValue(ctx, yesKey, yes)
}

func GetYes(ctx context.Context) bool {
	if v, ok := ctx.Value(yesKey).(bool); ok {
		return v
	}
	return false
}

func WithLimit(ctx context.Context, limit int) context.Context {
	return context.WithValue(ctx, limitKey, limit)
}

func GetLimit(ctx context.Context) int {
	if v, ok := ctx.Value(limitKey).(int); ok {
		return v
	}
	return 0
}

func WithSortBy(ctx context.Context, sortBy string) context.Context {
	return context.WithValue(ctx, sortByKey, sortBy)
}

func GetSortBy(ctx context.Context) string {
	if v, ok := ctx.Value(sortByKey).(string); ok {
		return v
	}
	return ""
}

func WithDesc(ctx context.Context, desc bool) context.Context {
	return context.WithValue(ctx, descKey, desc)
}

func GetDesc(ctx context.Context) bool {
	if v, ok := ctx.Value(descKey).(bool); ok {
		return v
	}
	return false
}

func FormatJSON(w io.Writer, data interface{}, query string) error {
	if query == "" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	// Apply jq query
	q, err := gojq.Parse(query)
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}

	iter := q.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(out))
	}
	return nil
}
```

**Step 4: Run tests**

Run: `go mod tidy && go test ./internal/outfmt/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/outfmt/format.go internal/outfmt/format_test.go go.mod go.sum
git commit -m "feat: add output formatting with jq support"
```

---

## Task 6: Auth Commands

**Files:**
- Create: `internal/cmd/auth.go`
- Create: `internal/cmd/auth_test.go`
- Create: `internal/auth/server.go`
- Create: `internal/auth/templates.go`

**Step 1: Write failing test for auth list**

```go
// internal/cmd/auth_test.go
package cmd

import (
	"bytes"
	"testing"
)

func TestAuthCmd_SubCommands(t *testing.T) {
	cmd := newAuthCmd()

	subCmds := []string{"login", "logout", "list", "switch", "status"}
	for _, name := range subCmds {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to exist", name)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/... -v -run TestAuthCmd`
Expected: FAIL with "newAuthCmd not defined"

**Step 3: Implement auth commands (skeleton)**

```go
// internal/cmd/auth.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Login, logout, and manage workspace credentials.",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthSwitchCmd())
	cmd.AddCommand(newAuthStatusCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Dub",
		Long:  "Opens a browser to enter your Dub API key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement browser-based auth
			fmt.Fprintln(cmd.OutOrStdout(), "Opening browser for authentication...")
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	var workspace string

	cmd := &cobra.Command{
		Use:   "logout [workspace]",
		Short: "Remove workspace credentials",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				workspace = args[0]
			}
			if workspace == "" {
				return fmt.Errorf("workspace name required")
			}

			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			if err := store.Delete(workspace); err != nil {
				return fmt.Errorf("failed to remove workspace: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed workspace: %s\n", workspace)
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace to remove")

	return cmd
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			creds, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(creds) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No workspaces configured. Run: dub auth login")
				return nil
			}

			for _, c := range creds {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s (added %s)\n", c.Name, c.CreatedAt.Format("2006-01-02"))
			}
			return nil
		},
	}
}

func newAuthSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <workspace>",
		Short: "Set default workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement default workspace config
			fmt.Fprintf(cmd.OutOrStdout(), "Switched to workspace: %s\n", args[0])
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			creds, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(creds) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated. Run: dub auth login")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Authenticated with %d workspace(s)\n", len(creds))
			return nil
		},
	}
}

func init() {
	// Register auth command with root in root.go
}
```

**Step 4: Wire auth command to root**

Add to `internal/cmd/root.go` in `NewRootCmd()`:

```go
cmd.AddCommand(newAuthCmd())
```

**Step 5: Run tests**

Run: `go test ./internal/cmd/... -v -run TestAuthCmd`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/cmd/auth.go internal/cmd/auth_test.go internal/cmd/root.go
git commit -m "feat: add auth commands (login, logout, list, switch, status)"
```

---

## Task 7: Browser-Based Auth Server

**Files:**
- Create: `internal/auth/server.go`
- Create: `internal/auth/templates.go`

**Step 1: Create auth server**

```go
// internal/auth/server.go
package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

var validWorkspaceName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type SetupResult struct {
	WorkspaceName string
	APIKey        string
	Error         error
}

type SetupServer struct {
	result        chan SetupResult
	shutdown      chan struct{}
	pendingResult *SetupResult
	pendingMu     sync.Mutex
	csrfToken     string
	store         secrets.Store
}

func NewSetupServer(store secrets.Store) (*SetupServer, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	return &SetupServer{
		result:    make(chan SetupResult, 1),
		shutdown:  make(chan struct{}),
		csrfToken: hex.EncodeToString(tokenBytes),
		store:     store,
	}, nil
}

func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		_ = server.Serve(listener)
	}()

	go func() {
		if err := openBrowser(baseURL); err != nil {
			slog.Info("failed to open browser", "url", baseURL)
		}
	}()

	fmt.Printf("Opening browser at %s\n", baseURL)
	fmt.Println("Waiting for authentication...")

	select {
	case result := <-s.result:
		_ = server.Shutdown(context.Background())
		return &result, nil
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil, ctx.Err()
	case <-s.shutdown:
		_ = server.Shutdown(context.Background())
		s.pendingMu.Lock()
		defer s.pendingMu.Unlock()
		if s.pendingResult != nil {
			return s.pendingResult, nil
		}
		return nil, fmt.Errorf("setup cancelled")
	}
}

func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{"CSRFToken": s.csrfToken}
	setSecurityHeaders(w)
	tmpl.Execute(w, data)
}

func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	var req struct {
		WorkspaceName string `json:"workspace_name"`
		APIKey        string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Invalid request"})
		return
	}

	req.WorkspaceName = strings.TrimSpace(req.WorkspaceName)
	req.APIKey = strings.TrimSpace(req.APIKey)

	if !validWorkspaceName.MatchString(req.WorkspaceName) {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": "Invalid workspace name"})
		return
	}

	if !strings.HasPrefix(req.APIKey, "dub_") {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": "API key must start with 'dub_'"})
		return
	}

	// Validate against Dub API
	client := api.NewClient(req.APIKey)
	resp, err := client.Get(r.Context(), "/links?limit=1")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": fmt.Sprintf("Connection failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiErr := api.ReadAPIError(resp)
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": apiErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Connection successful!"})
}

func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	var req struct {
		WorkspaceName string `json:"workspace_name"`
		APIKey        string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "Invalid request"})
		return
	}

	req.WorkspaceName = strings.TrimSpace(req.WorkspaceName)
	req.APIKey = strings.TrimSpace(req.APIKey)

	err := s.store.Set(req.WorkspaceName, secrets.Credentials{
		Name:   req.WorkspaceName,
		APIKey: req.APIKey,
	})
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": fmt.Sprintf("Failed to save: %v", err)})
		return
	}

	s.pendingMu.Lock()
	s.pendingResult = &SetupResult{
		WorkspaceName: req.WorkspaceName,
		APIKey:        req.APIKey,
	}
	s.pendingMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "workspace_name": req.WorkspaceName})
}

func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	s.pendingMu.Lock()
	workspaceName := ""
	if s.pendingResult != nil {
		workspaceName = s.pendingResult.WorkspaceName
	}
	s.pendingMu.Unlock()

	data := map[string]string{"WorkspaceName": workspaceName, "CSRFToken": s.csrfToken}
	setSecurityHeaders(w)
	tmpl.Execute(w, data)
}

func (s *SetupServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	s.pendingMu.Lock()
	if s.pendingResult != nil {
		s.result <- *s.pendingResult
	}
	s.pendingMu.Unlock()
	close(s.shutdown)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
```

**Step 2: Create templates**

```go
// internal/auth/templates.go
package auth

const setupTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Dub CLI Setup</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0a0a0a; color: #fff; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .container { background: #1a1a1a; border-radius: 12px; padding: 40px; max-width: 420px; width: 100%; }
        h1 { font-size: 24px; margin-bottom: 8px; }
        p { color: #888; margin-bottom: 24px; }
        label { display: block; font-size: 14px; margin-bottom: 8px; color: #ccc; }
        input { width: 100%; padding: 12px; border: 1px solid #333; border-radius: 8px; background: #0a0a0a; color: #fff; font-size: 16px; margin-bottom: 16px; }
        input:focus { outline: none; border-color: #6366f1; }
        button { width: 100%; padding: 12px; border: none; border-radius: 8px; background: #6366f1; color: #fff; font-size: 16px; cursor: pointer; }
        button:hover { background: #5558e3; }
        button:disabled { opacity: 0.5; cursor: not-allowed; }
        .error { color: #ef4444; font-size: 14px; margin-bottom: 16px; }
        .success { color: #22c55e; font-size: 14px; margin-bottom: 16px; }
        .help { font-size: 12px; color: #666; margin-top: 16px; }
        .help a { color: #6366f1; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Dub CLI Setup</h1>
        <p>Enter your Dub API credentials</p>
        <form id="form">
            <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
            <label>Workspace Name</label>
            <input type="text" id="workspace_name" placeholder="my-workspace" required pattern="[a-zA-Z0-9_-]+">
            <label>API Key</label>
            <input type="password" id="api_key" placeholder="dub_xxxxxx" required>
            <div id="message"></div>
            <button type="button" id="validate">Test Connection</button>
            <button type="submit" id="submit" style="display:none; margin-top:8px; background:#22c55e;">Save & Continue</button>
        </form>
        <p class="help">Get your API key from <a href="https://app.dub.co/settings/tokens" target="_blank">Dub Settings</a></p>
    </div>
    <script>
        const csrfToken = '{{.CSRFToken}}';
        const form = document.getElementById('form');
        const validateBtn = document.getElementById('validate');
        const submitBtn = document.getElementById('submit');
        const message = document.getElementById('message');

        validateBtn.onclick = async () => {
            message.className = '';
            message.textContent = 'Testing connection...';
            validateBtn.disabled = true;

            try {
                const resp = await fetch('/validate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({
                        workspace_name: document.getElementById('workspace_name').value,
                        api_key: document.getElementById('api_key').value
                    })
                });
                const data = await resp.json();
                if (data.success) {
                    message.className = 'success';
                    message.textContent = data.message;
                    submitBtn.style.display = 'block';
                } else {
                    message.className = 'error';
                    message.textContent = data.error;
                }
            } catch (e) {
                message.className = 'error';
                message.textContent = 'Connection failed';
            }
            validateBtn.disabled = false;
        };

        form.onsubmit = async (e) => {
            e.preventDefault();
            submitBtn.disabled = true;
            message.textContent = 'Saving...';

            try {
                const resp = await fetch('/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({
                        workspace_name: document.getElementById('workspace_name').value,
                        api_key: document.getElementById('api_key').value
                    })
                });
                const data = await resp.json();
                if (data.success) {
                    window.location.href = '/success';
                } else {
                    message.className = 'error';
                    message.textContent = data.error;
                }
            } catch (e) {
                message.className = 'error';
                message.textContent = 'Save failed';
            }
            submitBtn.disabled = false;
        };
    </script>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Dub CLI - Success</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0a0a0a; color: #fff; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .container { background: #1a1a1a; border-radius: 12px; padding: 40px; max-width: 420px; width: 100%; text-align: center; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        h1 { font-size: 24px; margin-bottom: 8px; }
        p { color: #888; margin-bottom: 24px; }
        code { background: #0a0a0a; padding: 8px 16px; border-radius: 8px; display: inline-block; margin: 8px 0; }
        button { padding: 12px 24px; border: none; border-radius: 8px; background: #6366f1; color: #fff; font-size: 16px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✓</div>
        <h1>Setup Complete!</h1>
        <p>Workspace <strong>{{.WorkspaceName}}</strong> is now configured.</p>
        <p>You can close this window and return to your terminal.</p>
        <code>dub links list</code>
        <br><br>
        <button onclick="complete()">Close</button>
    </div>
    <script>
        async function complete() {
            await fetch('/complete', {
                method: 'POST',
                headers: { 'X-CSRF-Token': '{{.CSRFToken}}' }
            });
            window.close();
        }
        // Auto-complete after 3 seconds
        setTimeout(complete, 3000);
    </script>
</body>
</html>`
```

**Step 3: Update auth login command to use server**

Update `internal/cmd/auth.go` `newAuthLoginCmd`:

```go
func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Dub",
		Long:  "Opens a browser to enter your Dub API key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.OpenDefault()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			server, err := auth.NewSetupServer(store)
			if err != nil {
				return err
			}

			result, err := server.Start(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully authenticated workspace: %s\n", result.WorkspaceName)
			return nil
		},
	}
}
```

Add import: `"github.com/salmonumbrella/dub-cli/internal/auth"`

**Step 4: Run tests**

Run: `go mod tidy && go build ./cmd/dub`
Expected: BUILD SUCCESS

**Step 5: Commit**

```bash
git add internal/auth/server.go internal/auth/templates.go internal/cmd/auth.go
git commit -m "feat: add browser-based auth server"
```

---

## Task 8: Links Commands

**Files:**
- Create: `internal/cmd/links.go`
- Create: `internal/cmd/links_test.go`

**Step 1: Write failing test**

```go
// internal/cmd/links_test.go
package cmd

import (
	"testing"
)

func TestLinksCmd_SubCommands(t *testing.T) {
	cmd := newLinksCmd()

	subCmds := []string{"create", "list", "get", "count", "update", "upsert", "delete", "bulk"}
	for _, name := range subCmds {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to exist", name)
		}
	}
}
```

**Step 2: Run test**

Run: `go test ./internal/cmd/... -v -run TestLinksCmd`
Expected: FAIL

**Step 3: Implement links commands**

```go
// internal/cmd/links.go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/outfmt"
	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

func newLinksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "links",
		Short: "Manage links",
		Long:  "Create, list, update, and delete short links.",
	}

	cmd.AddCommand(newLinksCreateCmd())
	cmd.AddCommand(newLinksListCmd())
	cmd.AddCommand(newLinksGetCmd())
	cmd.AddCommand(newLinksCountCmd())
	cmd.AddCommand(newLinksUpdateCmd())
	cmd.AddCommand(newLinksUpsertCmd())
	cmd.AddCommand(newLinksDeleteCmd())
	cmd.AddCommand(newLinksBulkCmd())

	return cmd
}

func newLinksCreateCmd() *cobra.Command {
	var urlFlag, key, domain string
	var tags []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new link",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"url": urlFlag,
			}
			if key != "" {
				body["key"] = key
			}
			if domain != "" {
				body["domain"] = domain
			}
			if len(tags) > 0 {
				body["tagIds"] = tags
			}

			resp, err := client.Post(cmd.Context(), "/links", body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&urlFlag, "url", "", "Destination URL (required)")
	cmd.Flags().StringVar(&key, "key", "", "Custom short link slug")
	cmd.Flags().StringVar(&domain, "domain", "", "Custom domain")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tag IDs")
	cmd.MarkFlagRequired("url")

	return cmd
}

func newLinksListCmd() *cobra.Command {
	var search, domain, tagId string
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all links",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if search != "" {
				params.Set("search", search)
			}
			if domain != "" {
				params.Set("domain", domain)
			}
			if tagId != "" {
				params.Set("tagId", tagId)
			}
			if page > 0 {
				params.Set("page", fmt.Sprintf("%d", page))
			}
			if pageSize > 0 {
				params.Set("pageSize", fmt.Sprintf("%d", pageSize))
			}
			if limit := outfmt.GetLimit(cmd.Context()); limit > 0 {
				params.Set("pageSize", fmt.Sprintf("%d", limit))
			}

			path := "/links"
			if len(params) > 0 {
				path += "?" + params.Encode()
			}

			resp, err := client.Get(cmd.Context(), path)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVarP(&search, "search", "s", "", "Search query")
	cmd.Flags().StringVar(&domain, "domain", "", "Filter by domain")
	cmd.Flags().StringVar(&tagId, "tag", "", "Filter by tag ID")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "Results per page")

	return cmd
}

func newLinksGetCmd() *cobra.Command {
	var linkId, externalId, domain, key string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get link details",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			params := url.Values{}
			if linkId != "" {
				params.Set("linkId", linkId)
			}
			if externalId != "" {
				params.Set("externalId", externalId)
			}
			if domain != "" && key != "" {
				params.Set("domain", domain)
				params.Set("key", key)
			}

			resp, err := client.Get(cmd.Context(), "/links/info?"+params.Encode())
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&linkId, "id", "", "Link ID")
	cmd.Flags().StringVar(&externalId, "external-id", "", "External ID")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain (use with --key)")
	cmd.Flags().StringVar(&key, "key", "", "Short link key (use with --domain)")

	return cmd
}

func newLinksCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count",
		Short: "Count links",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Get(cmd.Context(), "/links/count")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}
}

func newLinksUpdateCmd() *cobra.Command {
	var linkId, urlFlag, key string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a link",
		RunE: func(cmd *cobra.Command, args []string) error {
			if linkId == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{}
			if urlFlag != "" {
				body["url"] = urlFlag
			}
			if key != "" {
				body["key"] = key
			}

			resp, err := client.Patch(cmd.Context(), "/links/"+linkId, body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&linkId, "id", "", "Link ID (required)")
	cmd.Flags().StringVar(&urlFlag, "url", "", "New destination URL")
	cmd.Flags().StringVar(&key, "key", "", "New short link slug")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newLinksUpsertCmd() *cobra.Command {
	var urlFlag, key, domain string

	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a link",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"url": urlFlag,
			}
			if key != "" {
				body["key"] = key
			}
			if domain != "" {
				body["domain"] = domain
			}

			resp, err := client.Put(cmd.Context(), "/links/upsert", body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&urlFlag, "url", "", "Destination URL (required)")
	cmd.Flags().StringVar(&key, "key", "", "Custom short link slug")
	cmd.Flags().StringVar(&domain, "domain", "", "Custom domain")
	cmd.MarkFlagRequired("url")

	return cmd
}

func newLinksDeleteCmd() *cobra.Command {
	var linkId string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a link",
		RunE: func(cmd *cobra.Command, args []string) error {
			if linkId == "" {
				return fmt.Errorf("--id is required")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			resp, err := client.Delete(cmd.Context(), "/links/"+linkId)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&linkId, "id", "", "Link ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newLinksBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk link operations",
	}

	cmd.AddCommand(newLinksBulkCreateCmd())
	cmd.AddCommand(newLinksBulkUpdateCmd())
	cmd.AddCommand(newLinksBulkDeleteCmd())

	return cmd
}

func newLinksBulkCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Bulk create links (reads JSON from stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}

			var links []map[string]interface{}
			if err := json.Unmarshal(data, &links); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			resp, err := client.Post(cmd.Context(), "/links/bulk", links)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}
}

func newLinksBulkUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Bulk update links (reads JSON from stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}

			var links []map[string]interface{}
			if err := json.Unmarshal(data, &links); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			resp, err := client.Patch(cmd.Context(), "/links/bulk", links)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}
}

func newLinksBulkDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Bulk delete links (reads JSON array of IDs from stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}

			var body map[string]interface{}
			if err := json.Unmarshal(data, &body); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			resp, err := client.Delete(cmd.Context(), "/links/bulk")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			return handleResponse(cmd, resp)
		},
	}
}

// Helper functions

func getClient(ctx context.Context) (*api.Client, error) {
	store, err := secrets.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	// TODO: Support workspace selection via flag/env
	creds, err := store.List()
	if err != nil {
		return nil, err
	}

	if len(creds) == 0 {
		return nil, fmt.Errorf("not authenticated. Run: dub auth login")
	}

	return api.NewClient(creds[0].APIKey), nil
}

func handleResponse(cmd *cobra.Command, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		apiErr := api.ParseAPIError(body)
		return fmt.Errorf("%s", apiErr.Error())
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not JSON, print raw
		fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}

	query := outfmt.GetQuery(cmd.Context())
	return outfmt.FormatJSON(cmd.OutOrStdout(), data, query)
}
```

Add import at top: `"net/http"`

**Step 4: Wire to root command**

Add to `NewRootCmd()` in `root.go`:

```go
cmd.AddCommand(newLinksCmd())
```

**Step 5: Run tests**

Run: `go test ./internal/cmd/... -v -run TestLinksCmd`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/cmd/links.go internal/cmd/links_test.go internal/cmd/root.go
git commit -m "feat: add links commands (CRUD + bulk)"
```

---

## Task 9: Remaining Resource Commands (Skeleton)

Create skeleton commands for all remaining resources. Each follows the same pattern as links.

**Files to create:**
- `internal/cmd/analytics.go`
- `internal/cmd/domains.go`
- `internal/cmd/partners.go`
- `internal/cmd/customers.go`
- `internal/cmd/commissions.go`
- `internal/cmd/track.go`
- `internal/cmd/tags.go`
- `internal/cmd/folders.go`
- `internal/cmd/events.go`
- `internal/cmd/version.go`
- `internal/cmd/completion.go`

These follow the exact same pattern. Implementation delegated to sub-agents for parallel work.

---

## Task 10: GoReleaser & CI

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/ci.yml`

**Step 1: Create GoReleaser config**

```yaml
# .goreleaser.yaml
version: 2

project_name: dub-cli

builds:
  - id: dub-darwin
    main: ./cmd/dub
    binary: dub
    env:
      - CGO_ENABLED=1
    goos:
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/salmonumbrella/dub-cli/internal/cmd.Version={{.Version}}
      - -X github.com/salmonumbrella/dub-cli/internal/cmd.Commit={{.ShortCommit}}

  - id: dub-other
    main: ./cmd/dub
    binary: dub
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/salmonumbrella/dub-cli/internal/cmd.Version={{.Version}}
      - -X github.com/salmonumbrella/dub-cli/internal/cmd.Commit={{.ShortCommit}}

archives:
  - id: default
    formats:
      - tar.gz
    format_overrides:
      - goos: windows
        formats:
          - zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
```

**Step 2: Create CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go test -v ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go build ./cmd/dub
```

**Step 3: Commit**

```bash
git add .goreleaser.yaml .github/workflows/ci.yml
git commit -m "ci: add goreleaser and github actions"
```

---

## Summary

**Core infrastructure (Tasks 1-5):** Project scaffolding, root command, secrets store, API client, output formatting

**Auth (Tasks 6-7):** Auth commands, browser-based setup server

**Resources (Tasks 8-9):** Links commands (complete), other resources (skeleton)

**Release (Task 10):** GoReleaser, GitHub Actions

**Parallel execution opportunities:**
- Tasks 3, 4, 5 can run in parallel (secrets, api, outfmt are independent)
- Task 9 resources can be implemented in parallel by multiple sub-agents
