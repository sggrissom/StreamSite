package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ErrCodeAuth        ErrorCode = "AUTH_ERROR"
	ErrCodeValidation  ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
	ErrCodeForbidden   ErrorCode = "FORBIDDEN"
	ErrCodeInternal    ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest  ErrorCode = "BAD_REQUEST"
	ErrCodeTooLarge    ErrorCode = "FILE_TOO_LARGE"
	ErrCodeInvalidType ErrorCode = "INVALID_FILE_TYPE"
)

// AppError represents a structured application error
type AppError struct {
	Code        ErrorCode `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	RequestPath string    `json:"request_path,omitempty"`
}

// Error implements the error interface
func (e AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrorResponse represents the JSON response for errors
type ErrorResponse struct {
	Error AppError `json:"error"`
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, details ...string) *AppError {
	err := &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// RespondWithError sends a standardized error response
func RespondWithError(w http.ResponseWriter, r *http.Request, err *AppError, statusCode int) {
	// Add request path to error for context
	if r != nil {
		err.RequestPath = r.URL.Path
	}

	// Log the error with context
	_, file, line, ok := runtime.Caller(1)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	data := map[string]interface{}{
		"caller": caller,
		"error":  err.Error(),
		"code":   err.Code,
	}

	if r != nil {
		data["method"] = r.Method
		data["path"] = r.URL.Path
		LogErrorWithRequest(r, LogCategorySystem, err.Error(), data)
	} else {
		LogErrorSimple(LogCategorySystem, err.Error(), data)
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Send JSON response
	response := ErrorResponse{
		Error: *err,
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		log.Printf("Failed to encode error response: %v", encodeErr)
	}
}

// RespondAuthError sends an authentication error response
func RespondAuthError(w http.ResponseWriter, r *http.Request, message string) {
	err := NewAppError(ErrCodeAuth, message)
	RespondWithError(w, r, err, http.StatusUnauthorized)
}

// RespondValidationError sends a validation error response
func RespondValidationError(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	err := NewAppError(ErrCodeValidation, message, details...)
	RespondWithError(w, r, err, http.StatusBadRequest)
}

// RespondNotFoundError sends a not found error response
func RespondNotFoundError(w http.ResponseWriter, r *http.Request, message string) {
	err := NewAppError(ErrCodeNotFound, message)
	RespondWithError(w, r, err, http.StatusNotFound)
}

// RespondForbiddenError sends a forbidden error response
func RespondForbiddenError(w http.ResponseWriter, r *http.Request, message string) {
	err := NewAppError(ErrCodeForbidden, message)
	RespondWithError(w, r, err, http.StatusForbidden)
}

// RespondInternalError sends an internal server error response
func RespondInternalError(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	err := NewAppError(ErrCodeInternal, message, details...)
	RespondWithError(w, r, err, http.StatusInternalServerError)
}

// RespondBadRequestError sends a bad request error response
func RespondBadRequestError(w http.ResponseWriter, r *http.Request, message string) {
	err := NewAppError(ErrCodeBadRequest, message)
	RespondWithError(w, r, err, http.StatusBadRequest)
}
