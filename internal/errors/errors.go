package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput      ErrorCode = "INVALID_INPUT"
	ErrCodeMissingRequired   ErrorCode = "MISSING_REQUIRED"
	ErrCodeInvalidFormat     ErrorCode = "INVALID_FORMAT"

	// Not found errors
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeResourceNotFound  ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodePageNotFound      ErrorCode = "PAGE_NOT_FOUND"

	// System errors
	ErrCodeInternal          ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase          ErrorCode = "DATABASE_ERROR"
	ErrCodeIO                ErrorCode = "IO_ERROR"
	ErrCodeNetwork           ErrorCode = "NETWORK_ERROR"
	ErrCodeSerialization     ErrorCode = "SERIALIZATION_ERROR"
	ErrCodeDecryption        ErrorCode = "DECRYPTION_ERROR"
	
	// Security errors
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	
	// Business logic errors
	ErrCodeBusinessLogic     ErrorCode = "BUSINESS_LOGIC_ERROR"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodePrecondition      ErrorCode = "PRECONDITION_FAILED"
)

// AppError represents an application error with context
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
	Stack   []string               `json:"-"`
	Cause   error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// IsUserError returns true if this error should be shown to users
func (e *AppError) IsUserError() bool {
	switch e.Code {
	case ErrCodeValidation, ErrCodeInvalidInput, ErrCodeMissingRequired,
		ErrCodeInvalidFormat, ErrCodeNotFound, ErrCodeResourceNotFound,
		ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeRateLimitExceeded,
		ErrCodeConflict, ErrCodePrecondition:
		return true
	default:
		return false
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetStack returns the stack trace as a string
func (e *AppError) GetStack() string {
	return strings.Join(e.Stack, "\n")
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Stack:   captureStack(),
	}
}

// Wrap wraps an existing error
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, preserve the original
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:    code,
			Message: message,
			Context: appErr.Context,
			Stack:   captureStack(),
			Cause:   appErr,
		}
	}
	
	return &AppError{
		Code:    code,
		Message: message,
		Stack:   captureStack(),
		Cause:   err,
	}
}

// Validation helpers
func Validation(field string, message string) *AppError {
	return New(ErrCodeValidation, message).
		WithContext("field", field)
}

func NotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource)).
		WithContext("resource", resource)
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return New(ErrCodeUnauthorized, message)
}

func Internal(message string) *AppError {
	return New(ErrCodeInternal, message)
}

func RateLimitExceeded(limit int) *AppError {
	return New(ErrCodeRateLimitExceeded, "Rate limit exceeded").
		WithContext("limit", limit)
}

// captureStack captures the current stack trace
func captureStack() []string {
	var stack []string
	
	// Skip the first few frames to get to the actual error location
	for i := 2; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		// Skip runtime and testing frames
		if strings.Contains(file, "runtime/") || strings.Contains(file, "testing/") {
			continue
		}
		
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	
	return stack
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetCode returns the error code if it's an AppError, otherwise returns ErrCodeInternal
func GetCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrCodeInternal
}