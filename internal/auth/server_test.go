// internal/auth/server_test.go
package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

// MockStore implements secrets.Store for testing
type MockStore struct {
	credentials map[string]secrets.Credentials
	setErr      error
	getErr      error
}

func NewMockStore() *MockStore {
	return &MockStore{
		credentials: make(map[string]secrets.Credentials),
	}
}

func (m *MockStore) Keys() ([]string, error) {
	keys := make([]string, 0, len(m.credentials))
	for k := range m.credentials {
		keys = append(keys, "workspace:"+k)
	}
	return keys, nil
}

func (m *MockStore) Set(name string, creds secrets.Credentials) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.credentials[name] = creds
	return nil
}

func (m *MockStore) Get(name string) (secrets.Credentials, error) {
	if m.getErr != nil {
		return secrets.Credentials{}, m.getErr
	}
	creds, ok := m.credentials[name]
	if !ok {
		return secrets.Credentials{}, nil
	}
	return creds, nil
}

func (m *MockStore) Delete(name string) error {
	delete(m.credentials, name)
	return nil
}

func (m *MockStore) List() ([]secrets.Credentials, error) {
	list := make([]secrets.Credentials, 0, len(m.credentials))
	for _, creds := range m.credentials {
		list = append(list, creds)
	}
	return list, nil
}

// Test CSRF token generation
func TestGenerateCSRFToken(t *testing.T) {
	token1, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected token length of 64, got %d", len(token1))
	}

	// Tokens should be unique
	token2, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token1 == token2 {
		t.Error("expected unique tokens, got duplicates")
	}
}

// Test CSRF token format (hex encoding)
func TestGenerateCSRFToken_Format(t *testing.T) {
	token, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should only contain hex characters
	for _, c := range token {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("token contains non-hex character: %c", c)
		}
	}
}

// Test NewSetupServer creates server with CSRF token
func TestNewSetupServer(t *testing.T) {
	store := NewMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if server.csrfToken == "" {
		t.Error("expected CSRF token to be set")
	}

	if server.store != store {
		t.Error("expected store to be set")
	}

	if server.doneChan == nil {
		t.Error("expected doneChan to be initialized")
	}
}

// Test API key format validation
func TestAPIKeyValidation(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		isValid bool
	}{
		{"valid key", "dub_abc123xyz", true},
		{"valid key with special chars", "dub_AbC123-_xyz", true},
		{"missing prefix", "abc123xyz", false},
		{"wrong prefix", "api_abc123", false},
		{"empty string", "", false},
		{"only prefix", "dub_", true}, // Prefix check passes, API validation would fail
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPrefix := strings.HasPrefix(tt.apiKey, "dub_")
			if hasPrefix != tt.isValid {
				t.Errorf("apiKey %q: hasPrefix=%v, want %v", tt.apiKey, hasPrefix, tt.isValid)
			}
		})
	}
}

// Test handleSetup serves the form
func TestHandleSetup(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	// Test GET with valid CSRF
	req := httptest.NewRequest(http.MethodGet, "/?csrf="+server.csrfToken, nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got %q", ct)
	}

	// Check that response contains CSRF token
	body := w.Body.String()
	if !strings.Contains(body, server.csrfToken) {
		t.Error("expected response to contain CSRF token")
	}
}

// Test handleSetup rejects invalid CSRF
func TestHandleSetup_InvalidCSRF(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodGet, "/?csrf=invalid_token", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// Test handleSetup rejects non-GET methods
func TestHandleSetup_WrongMethod(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodPost, "/?csrf="+server.csrfToken, nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// Test handleValidate with valid CSRF but missing API key
func TestHandleValidate_MissingAPIKey(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", server.csrfToken)

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "API key is required" {
		t.Errorf("expected 'API key is required' error, got %q", resp["error"])
	}
}

// Test handleValidate with invalid API key prefix
func TestHandleValidate_InvalidPrefix(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", server.csrfToken)
	form.Set("api_key", "invalid_key")

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "API key must start with 'dub_'" {
		t.Errorf("expected prefix error, got %q", resp["error"])
	}
}

// Test handleValidate rejects invalid CSRF
func TestHandleValidate_InvalidCSRF(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", "wrong_token")
	form.Set("api_key", "dub_test123")

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// Test handleValidate rejects non-POST methods
func TestHandleValidate_WrongMethod(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// Test handleSubmit with missing workspace
func TestHandleSubmit_MissingWorkspace(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", server.csrfToken)
	form.Set("api_key", "dub_test123")

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "Workspace name is required" {
		t.Errorf("expected 'Workspace name is required' error, got %q", resp["error"])
	}
}

// Test handleSubmit with missing API key
func TestHandleSubmit_MissingAPIKey(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", server.csrfToken)
	form.Set("workspace", "my-workspace")

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "API key is required" {
		t.Errorf("expected 'API key is required' error, got %q", resp["error"])
	}
}

// Test handleSubmit with invalid API key prefix
func TestHandleSubmit_InvalidPrefix(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", server.csrfToken)
	form.Set("workspace", "my-workspace")
	form.Set("api_key", "invalid_key")

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "API key must start with 'dub_'" {
		t.Errorf("expected prefix error, got %q", resp["error"])
	}
}

// Test handleSubmit rejects invalid CSRF
func TestHandleSubmit_InvalidCSRF(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	form := url.Values{}
	form.Set("csrf_token", "wrong_token")
	form.Set("workspace", "my-workspace")
	form.Set("api_key", "dub_test123")

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// Test handleSubmit rejects non-POST methods
func TestHandleSubmit_WrongMethod(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodGet, "/submit", nil)
	w := httptest.NewRecorder()

	server.handleSubmit(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// Test handleSuccess serves success page
func TestHandleSuccess(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodGet, "/success?csrf="+server.csrfToken+"&workspace=test-ws", nil)
	w := httptest.NewRecorder()

	server.handleSuccess(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got %q", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "test-ws") {
		t.Error("expected response to contain workspace name")
	}
}

// Test handleSuccess rejects invalid CSRF
func TestHandleSuccess_InvalidCSRF(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodGet, "/success?csrf=invalid_token", nil)
	w := httptest.NewRecorder()

	server.handleSuccess(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// Test handleSuccess rejects non-GET methods
func TestHandleSuccess_WrongMethod(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodPost, "/success?csrf="+server.csrfToken, nil)
	w := httptest.NewRecorder()

	server.handleSuccess(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// Test handleComplete signals completion
func TestHandleComplete(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodPost, "/complete?csrf="+server.csrfToken, nil)
	w := httptest.NewRecorder()

	server.handleComplete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}

	// Verify channel is closed
	select {
	case <-server.doneChan:
		// Expected - channel is closed
	default:
		t.Error("expected doneChan to be closed")
	}
}

// Test handleComplete rejects invalid CSRF
func TestHandleComplete_InvalidCSRF(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodPost, "/complete?csrf=invalid_token", nil)
	w := httptest.NewRecorder()

	server.handleComplete(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// Test handleComplete rejects non-POST methods
func TestHandleComplete_WrongMethod(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	req := httptest.NewRequest(http.MethodGet, "/complete?csrf="+server.csrfToken, nil)
	w := httptest.NewRecorder()

	server.handleComplete(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

// Test handleComplete is idempotent (multiple calls don't panic)
func TestHandleComplete_Idempotent(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	// Call multiple times - should not panic
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/complete?csrf="+server.csrfToken, nil)
		w := httptest.NewRecorder()
		server.handleComplete(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("call %d: expected status 200, got %d", i+1, w.Code)
		}
	}
}

// Test setResult is thread-safe
func TestSetResult_ThreadSafe(t *testing.T) {
	store := NewMockStore()
	server, _ := NewSetupServer(store)

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			server.setResult(&SetupResult{
				WorkspaceName: "workspace",
				APIKey:        "dub_test",
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have a result set (no race panic)
	server.mu.Lock()
	if server.result == nil {
		t.Error("expected result to be set")
	}
	server.mu.Unlock()
}

// Test writeJSON helper
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	writeJSON(w, http.StatusCreated, map[string]string{"message": "created"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["message"] != "created" {
		t.Errorf("expected message 'created', got %q", resp["message"])
	}
}

// Test SetupResult struct
func TestSetupResult(t *testing.T) {
	result := SetupResult{
		WorkspaceName: "test-workspace",
		APIKey:        "dub_abc123",
		Error:         nil,
	}

	if result.WorkspaceName != "test-workspace" {
		t.Errorf("expected workspace 'test-workspace', got %q", result.WorkspaceName)
	}

	if result.APIKey != "dub_abc123" {
		t.Errorf("expected API key 'dub_abc123', got %q", result.APIKey)
	}

	if result.Error != nil {
		t.Errorf("expected nil error, got %v", result.Error)
	}
}

// Test templates contain expected elements
func TestSetupTemplate_ContainsElements(t *testing.T) {
	expectedElements := []string{
		"Connect to Dub",
		"Workspace Name",
		"API Key",
		"Save & Connect",
		"csrf_token",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(setupTemplate, elem) {
			t.Errorf("setup template should contain %q", elem)
		}
	}
}

func TestSuccessTemplate_ContainsElements(t *testing.T) {
	expectedElements := []string{
		"You're all set!",
		"Workspace",
		"links list",
		"links create",
		"dub --help",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(successTemplate, elem) {
			t.Errorf("success template should contain %q", elem)
		}
	}
}

// Test MockStore implements secrets.Store interface correctly
func TestMockStore_Interface(t *testing.T) {
	var _ secrets.Store = NewMockStore()
}

func TestMockStore_SetAndGet(t *testing.T) {
	store := NewMockStore()

	creds := secrets.Credentials{
		Name:      "test-ws",
		APIKey:    "dub_test123",
		CreatedAt: time.Now(),
	}

	err := store.Set("test-ws", creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := store.Get("test-ws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Name != creds.Name {
		t.Errorf("expected name %q, got %q", creds.Name, got.Name)
	}

	if got.APIKey != creds.APIKey {
		t.Errorf("expected API key %q, got %q", creds.APIKey, got.APIKey)
	}
}

func TestMockStore_Delete(t *testing.T) {
	store := NewMockStore()

	creds := secrets.Credentials{
		Name:   "test-ws",
		APIKey: "dub_test123",
	}

	if err := store.Set("test-ws", creds); err != nil {
		t.Fatalf("failed to set credentials: %v", err)
	}
	if err := store.Delete("test-ws"); err != nil {
		t.Fatalf("failed to delete credentials: %v", err)
	}

	got, _ := store.Get("test-ws")
	if got.Name != "" {
		t.Error("expected credential to be deleted")
	}
}

func TestMockStore_Keys(t *testing.T) {
	store := NewMockStore()

	if err := store.Set("ws1", secrets.Credentials{Name: "ws1", APIKey: "dub_1"}); err != nil {
		t.Fatalf("failed to set ws1: %v", err)
	}
	if err := store.Set("ws2", secrets.Credentials{Name: "ws2", APIKey: "dub_2"}); err != nil {
		t.Fatalf("failed to set ws2: %v", err)
	}

	keys, err := store.Keys()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestMockStore_List(t *testing.T) {
	store := NewMockStore()

	if err := store.Set("ws1", secrets.Credentials{Name: "ws1", APIKey: "dub_1"}); err != nil {
		t.Fatalf("failed to set ws1: %v", err)
	}
	if err := store.Set("ws2", secrets.Credentials{Name: "ws2", APIKey: "dub_2"}); err != nil {
		t.Fatalf("failed to set ws2: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 credentials, got %d", len(list))
	}
}
