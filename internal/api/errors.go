package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
		// Check if response looks like HTML (not a JSON API response)
		bodyStr := string(body)
		if len(bodyStr) > 0 && (bodyStr[0] == '<' || strings.HasPrefix(strings.TrimSpace(bodyStr), "<!")) {
			return &APIError{
				Code:    "not_found",
				Message: "Resource not found or endpoint does not exist",
			}
		}
		return &APIError{Message: bodyStr}
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
