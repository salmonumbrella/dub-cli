// internal/auth/server.go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/dub-cli/internal/api"
	"github.com/salmonumbrella/dub-cli/internal/secrets"
)

// SetupResult contains the result of the authentication flow.
type SetupResult struct {
	WorkspaceName string
	APIKey        string
	Error         error
}

// SetupServer handles browser-based authentication.
type SetupServer struct {
	store     secrets.Store
	csrfToken string
	listener  net.Listener
	server    *http.Server

	mu       sync.Mutex
	result   *SetupResult
	doneChan chan struct{}
}

// NewSetupServer creates a new setup server.
func NewSetupServer(store secrets.Store) (*SetupServer, error) {
	token, err := generateCSRFToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	return &SetupServer{
		store:     store,
		csrfToken: token,
		doneChan:  make(chan struct{}),
	}, nil
}

// Start launches the HTTP server and opens the browser.
// It blocks until authentication is complete or the context is cancelled.
func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
	// Bind to a random port on localhost
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)

	s.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in background
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.setResult(&SetupResult{Error: err})
		}
	}()

	// Open browser
	url := fmt.Sprintf("http://%s/?csrf=%s", listener.Addr().String(), s.csrfToken)
	if err := openBrowser(url); err != nil {
		fmt.Printf("Please open this URL in your browser:\n  %s\n\n", url)
	}

	// Wait for completion or context cancellation
	select {
	case <-s.doneChan:
		// Give browser time to load success page
		time.Sleep(500 * time.Millisecond)
	case <-ctx.Done():
		s.setResult(&SetupResult{Error: ctx.Err()})
	}

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.server.Shutdown(shutdownCtx)

	s.mu.Lock()
	result := s.result
	s.mu.Unlock()

	if result == nil {
		return nil, fmt.Errorf("authentication cancelled")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return result, nil
}

// handleSetup serves the main setup form.
func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	csrf := r.URL.Query().Get("csrf")
	if csrf != s.csrfToken {
		http.Error(w, "Invalid request", http.StatusForbidden)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := struct {
		CSRFToken string
	}{
		CSRFToken: s.csrfToken,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// handleValidate tests the API key without saving.
func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid form data"})
		return
	}

	csrf := r.FormValue("csrf_token")
	if csrf != s.csrfToken {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Invalid request"})
		return
	}

	apiKey := strings.TrimSpace(r.FormValue("api_key"))
	if apiKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API key is required"})
		return
	}

	if !strings.HasPrefix(apiKey, "dub_") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API key must start with 'dub_'"})
		return
	}

	// Test the API key
	if err := validateAPIKey(r.Context(), apiKey); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "valid"})
}

// handleSubmit saves the credentials.
func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid form data"})
		return
	}

	csrf := r.FormValue("csrf_token")
	if csrf != s.csrfToken {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Invalid request"})
		return
	}

	workspace := strings.TrimSpace(r.FormValue("workspace"))
	apiKey := strings.TrimSpace(r.FormValue("api_key"))

	if workspace == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Workspace name is required"})
		return
	}

	if apiKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API key is required"})
		return
	}

	if !strings.HasPrefix(apiKey, "dub_") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API key must start with 'dub_'"})
		return
	}

	// Validate the API key before saving
	if err := validateAPIKey(r.Context(), apiKey); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	// Save to keyring
	creds := secrets.Credentials{
		Name:      workspace,
		APIKey:    apiKey,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.store.Set(workspace, creds); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to save credentials: %v", err)})
		return
	}

	// Set result for CLI
	s.setResult(&SetupResult{
		WorkspaceName: workspace,
		APIKey:        apiKey,
	})

	// Return success with redirect URL
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "saved",
		"redirect": fmt.Sprintf("/success?csrf=%s&workspace=%s", s.csrfToken, workspace),
	})
}

// handleSuccess shows the success page.
func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	csrf := r.URL.Query().Get("csrf")
	if csrf != s.csrfToken {
		http.Error(w, "Invalid request", http.StatusForbidden)
		return
	}

	workspace := r.URL.Query().Get("workspace")

	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Workspace string
		CSRFToken string
	}{
		Workspace: workspace,
		CSRFToken: s.csrfToken,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// handleComplete signals the CLI that authentication is done.
func (s *SetupServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	csrf := r.URL.Query().Get("csrf")
	if csrf != s.csrfToken {
		http.Error(w, "Invalid request", http.StatusForbidden)
		return
	}

	// Signal completion
	select {
	case <-s.doneChan:
		// Already closed
	default:
		close(s.doneChan)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *SetupServer) setResult(result *SetupResult) {
	s.mu.Lock()
	s.result = result
	s.mu.Unlock()
}

// validateAPIKey tests the API key against the Dub API.
func validateAPIKey(ctx context.Context, apiKey string) error {
	client := api.NewClient(apiKey)
	resp, err := client.Get(ctx, "/links?limit=1")
	if err != nil {
		return fmt.Errorf("failed to connect to Dub API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode == 403 {
		return fmt.Errorf("API key does not have permission to access links")
	}
	if resp.StatusCode >= 400 {
		apiErr := api.ReadAPIError(resp)
		return fmt.Errorf("API error: %s", apiErr.Message)
	}

	return nil
}

// generateCSRFToken creates a random token for CSRF protection.
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// writeJSON writes a JSON response with proper encoding.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// openBrowser opens the default browser to the specified URL.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}
