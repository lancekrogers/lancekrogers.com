package contact

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {
	v := NewValidator()
	
	t.Run("valid request", func(t *testing.T) {
		req := &ContactRequest{
			Name:    "John Doe",
			Email:   "john@example.com",
			Company: "ACME Corp",
			Message: "I would like to discuss a project with you.",
		}
		
		errors := v.Validate(req)
		assert.Empty(t, errors)
	})
	
	t.Run("missing required fields", func(t *testing.T) {
		req := &ContactRequest{}
		
		errors := v.Validate(req)
		assert.Len(t, errors, 3) // name, email, message
		
		// Check specific errors
		errMap := make(map[string]string)
		for _, err := range errors {
			errMap[err.Field] = err.Message
		}
		
		assert.Contains(t, errMap["name"], "required")
		assert.Contains(t, errMap["email"], "required")
		assert.Contains(t, errMap["message"], "required")
	})
	
	t.Run("invalid name", func(t *testing.T) {
		tests := []struct {
			name    string
			value   string
			wantErr bool
		}{
			{"valid name", "John Doe", false},
			{"with hyphen", "Mary-Jane", false},
			{"with apostrophe", "O'Brien", false},
			{"too short", "J", true},
			{"with numbers", "John123", true},
			{"with special chars", "John@Doe", true},
			{"empty after trim", "   ", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := &ContactRequest{
					Name:    tt.value,
					Email:   "test@example.com",
					Message: "This is a test message.",
				}
				
				errors := v.Validate(req)
				if tt.wantErr {
					assert.NotEmpty(t, errors)
					assert.Equal(t, "name", errors[0].Field)
				} else {
					assert.Empty(t, errors)
				}
			})
		}
	})
	
	t.Run("invalid email", func(t *testing.T) {
		tests := []struct {
			email   string
			wantErr bool
		}{
			{"valid@example.com", false},
			{"user.name@example.com", false},
			{"user+tag@example.co.uk", false},
			{"invalid", true},
			{"@example.com", true},
			{"user@", true},
			{"user space@example.com", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.email, func(t *testing.T) {
				req := &ContactRequest{
					Name:    "Test User",
					Email:   tt.email,
					Message: "This is a test message.",
				}
				
				errors := v.Validate(req)
				if tt.wantErr {
					assert.NotEmpty(t, errors)
					found := false
					for _, err := range errors {
						if err.Field == "email" {
							found = true
							break
						}
					}
					assert.True(t, found, "expected email error")
				} else {
					assert.Empty(t, errors)
				}
			})
		}
	})
	
	t.Run("optional company field", func(t *testing.T) {
		// Empty company should be valid
		req := &ContactRequest{
			Name:    "John Doe",
			Email:   "john@example.com",
			Message: "Test message for validation.",
		}
		errors := v.Validate(req)
		assert.Empty(t, errors)
		
		// Valid company
		req.Company = "ACME & Sons, Inc."
		errors = v.Validate(req)
		assert.Empty(t, errors)
		
		// Invalid company
		req.Company = "Company<script>"
		errors = v.Validate(req)
		assert.NotEmpty(t, errors)
		assert.Equal(t, "company", errors[0].Field)
	})
	
	t.Run("message validation", func(t *testing.T) {
		// Too short
		req := &ContactRequest{
			Name:    "John Doe",
			Email:   "john@example.com",
			Message: "Hi",
		}
		errors := v.Validate(req)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Message, "10 characters")
		
		// Too long
		req.Message = strings.Repeat("a", 5001)
		errors = v.Validate(req)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Message, "5000 characters")
		
		// With special characters and newlines
		req.Message = "Hello,\n\nI'd like to discuss: \n- Item 1\n- Item 2\n\nThanks!"
		errors = v.Validate(req)
		assert.Empty(t, errors)
		
		// Unicode characters
		req.Message = "Hello! 你好 🎉 I'd like to discuss a project."
		errors = v.Validate(req)
		assert.Empty(t, errors)
	})
	
	t.Run("trimming whitespace", func(t *testing.T) {
		req := &ContactRequest{
			Name:    "  John Doe  ",
			Email:   "  JOHN@EXAMPLE.COM  ",
			Company: "  ACME Corp  ",
			Message: "  Test message with spaces.  ",
		}
		
		errors := v.Validate(req)
		assert.Empty(t, errors)
		
		// Check that values were trimmed and normalized
		assert.Equal(t, "John Doe", req.Name)
		assert.Equal(t, "john@example.com", req.Email) // Also lowercased
		assert.Equal(t, "ACME Corp", req.Company)
		assert.Equal(t, "Test message with spaces.", req.Message)
	})
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"John & Jane", "John &amp; Jane"},
		{`"Hello" 'World'`, "&quot;Hello&quot; &#39;World&#39;"},
		{"<img src=x onerror=alert(1)>", "&lt;img src=x onerror=alert(1)&gt;"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}