package hetzner

import (
	"fmt"
	"net/http"
)

// APIError represents a typed API error from Hetzner Cloud API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("Hetzner API error (status=%d, request_id=%s): %s", e.StatusCode, e.RequestID, e.Message)
	}
	return fmt.Sprintf("Hetzner API error (status=%d): %s", e.StatusCode, e.Message)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		return e.StatusCode == t.StatusCode
	}
	return false
}

// Predefined API errors
var (
	ErrUnauthorized = &APIError{
		StatusCode: http.StatusUnauthorized,
		Message:    "Unauthorized: Invalid or missing API token",
	}

	ErrForbidden = &APIError{
		StatusCode: http.StatusForbidden,
		Message:    "Forbidden: Insufficient permissions for this operation",
	}

	ErrRateLimited = &APIError{
		StatusCode: http.StatusTooManyRequests,
		Message:    "Rate limited: Too many requests, please try again later",
	}

	ErrServerError = &APIError{
		StatusCode: 0, // Will be set dynamically for 5xx errors
		Message:    "Hetzner server error",
	}

	ErrNotFound = &APIError{
		StatusCode: http.StatusNotFound,
		Message:    "Not found: The requested resource was not found",
	}

	ErrBadRequest = &APIError{
		StatusCode: http.StatusBadRequest,
		Message:    "Bad request: The request was invalid or cannot be served",
	}
)

// NewAPIError creates a new APIError with the given status code and message
func NewAPIError(statusCode int, message string, requestID string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		RequestID:  requestID,
	}
}

// NewAPIErrorWithWrap creates a new APIError wrapping an existing error
func NewAPIErrorWithWrap(statusCode int, message string, requestID string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		RequestID:  requestID,
		Err:        err,
	}
}

// IsAPIError checks if the error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError returns the APIError if err is an APIError, nil otherwise
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}

// HTTPErrorToAPIError converts HTTP status codes to typed API errors
func HTTPErrorToAPIError(statusCode int, requestID string) *APIError {
	switch statusCode {
	case http.StatusUnauthorized:
		return NewAPIError(statusCode, "Unauthorized: Invalid or missing API token", requestID)
	case http.StatusForbidden:
		return NewAPIError(statusCode, "Forbidden: Insufficient permissions for this operation", requestID)
	case http.StatusTooManyRequests:
		return NewAPIError(statusCode, "Rate limited: Too many requests, please try again later", requestID)
	case http.StatusNotFound:
		return NewAPIError(statusCode, "Not found: The requested resource was not found", requestID)
	case http.StatusBadRequest:
		return NewAPIError(statusCode, "Bad request: The request was invalid or cannot be served", requestID)
	default:
		if statusCode >= 500 {
			return NewAPIError(statusCode, fmt.Sprintf("Server error: Hetzner API returned status %d", statusCode), requestID)
		}
		return NewAPIError(statusCode, fmt.Sprintf("HTTP error: Hetzner API returned status %d", statusCode), requestID)
	}
}

// IsRetryableError returns true if the error is retryable (rate limit or server error)
func IsRetryableError(err error) bool {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.StatusCode == http.StatusTooManyRequests || (apiErr.StatusCode >= 500 && apiErr.StatusCode < 600)
	}
	return false
}

// IsAuthError returns true if the error is authentication/authorization related
func IsAuthError(err error) bool {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden
	}
	return false
}

// IsClientError returns true if the error is a client error (4xx)
func IsClientError(err error) bool {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
	}
	return false
}

// IsServerError returns true if the error is a server error (5xx)
func IsServerError(err error) bool {
	if apiErr := GetAPIError(err); apiErr != nil {
		return apiErr.StatusCode >= 500 && apiErr.StatusCode < 600
	}
	return false
}
