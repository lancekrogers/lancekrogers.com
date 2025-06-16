package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("create new error", func(t *testing.T) {
		err := New(ErrCodeValidation, "validation failed")
		
		assert.Equal(t, ErrCodeValidation, err.Code)
		assert.Equal(t, "validation failed", err.Message)
		assert.Contains(t, err.Error(), "[VALIDATION_ERROR]")
		assert.Contains(t, err.Error(), "validation failed")
		assert.NotEmpty(t, err.Stack)
	})
	
	t.Run("wrap existing error", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		wrappedErr := Wrap(originalErr, ErrCodeInternal, "something went wrong")
		
		assert.Equal(t, ErrCodeInternal, wrappedErr.Code)
		assert.Equal(t, "something went wrong", wrappedErr.Message)
		assert.Equal(t, originalErr, wrappedErr.Cause)
		assert.Contains(t, wrappedErr.Error(), "original error")
	})
	
	t.Run("wrap nil error", func(t *testing.T) {
		wrappedErr := Wrap(nil, ErrCodeInternal, "should be nil")
		assert.Nil(t, wrappedErr)
	})
	
	t.Run("wrap app error", func(t *testing.T) {
		originalErr := New(ErrCodeValidation, "validation error").
			WithContext("field", "email")
		wrappedErr := Wrap(originalErr, ErrCodeInternal, "request failed")
		
		assert.Equal(t, ErrCodeInternal, wrappedErr.Code)
		assert.Equal(t, originalErr, wrappedErr.Cause)
		// Context should be preserved
		assert.Equal(t, "email", originalErr.Context["field"])
	})
	
	t.Run("with context", func(t *testing.T) {
		err := New(ErrCodeValidation, "invalid input").
			WithContext("field", "email").
			WithContext("value", "not-an-email")
		
		assert.Equal(t, "email", err.Context["field"])
		assert.Equal(t, "not-an-email", err.Context["value"])
	})
	
	t.Run("is user error", func(t *testing.T) {
		userErrors := []ErrorCode{
			ErrCodeValidation,
			ErrCodeInvalidInput,
			ErrCodeNotFound,
			ErrCodeUnauthorized,
			ErrCodeForbidden,
			ErrCodeRateLimitExceeded,
		}
		
		for _, code := range userErrors {
			err := New(code, "test")
			assert.True(t, err.IsUserError(), "Code %s should be a user error", code)
		}
		
		systemErrors := []ErrorCode{
			ErrCodeInternal,
			ErrCodeDatabase,
			ErrCodeIO,
			ErrCodeNetwork,
		}
		
		for _, code := range systemErrors {
			err := New(code, "test")
			assert.False(t, err.IsUserError(), "Code %s should not be a user error", code)
		}
	})
	
	t.Run("helper functions", func(t *testing.T) {
		// Validation error
		err := Validation("email", "invalid email format")
		assert.Equal(t, ErrCodeValidation, err.Code)
		assert.Equal(t, "email", err.Context["field"])
		
		// Not found error
		err = NotFound("user")
		assert.Equal(t, ErrCodeNotFound, err.Code)
		assert.Equal(t, "user", err.Context["resource"])
		assert.Contains(t, err.Message, "user not found")
		
		// Unauthorized error
		err = Unauthorized("")
		assert.Equal(t, ErrCodeUnauthorized, err.Code)
		assert.Equal(t, "Unauthorized access", err.Message)
		
		err = Unauthorized("custom message")
		assert.Equal(t, "custom message", err.Message)
		
		// Internal error
		err = Internal("system failure")
		assert.Equal(t, ErrCodeInternal, err.Code)
		assert.Equal(t, "system failure", err.Message)
		
		// Rate limit error
		err = RateLimitExceeded(100)
		assert.Equal(t, ErrCodeRateLimitExceeded, err.Code)
		assert.Equal(t, 100, err.Context["limit"])
	})
	
	t.Run("is app error", func(t *testing.T) {
		appErr := New(ErrCodeValidation, "test")
		stdErr := fmt.Errorf("standard error")
		
		assert.True(t, IsAppError(appErr))
		assert.False(t, IsAppError(stdErr))
	})
	
	t.Run("get code", func(t *testing.T) {
		appErr := New(ErrCodeValidation, "test")
		stdErr := fmt.Errorf("standard error")
		
		assert.Equal(t, ErrCodeValidation, GetCode(appErr))
		assert.Equal(t, ErrCodeInternal, GetCode(stdErr))
	})
	
	t.Run("stack trace", func(t *testing.T) {
		err := New(ErrCodeInternal, "test error")
		stack := err.GetStack()
		
		assert.NotEmpty(t, stack)
		assert.Contains(t, stack, "errors_test.go")
	})
}