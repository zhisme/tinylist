package worker_test

import (
	"testing"

	"github.com/zhisme/tinylist/internal/worker"
)

func TestReplaceTemplateVars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		userName string
		email    string
		expected string
	}{
		{
			name:     "replace name only",
			text:     "Hello {{name}}!",
			userName: "John",
			email:    "john@example.com",
			expected: "Hello John!",
		},
		{
			name:     "replace email only",
			text:     "Your email is {{email}}",
			userName: "John",
			email:    "john@example.com",
			expected: "Your email is john@example.com",
		},
		{
			name:     "replace both",
			text:     "Hi {{name}}, we'll contact you at {{email}}",
			userName: "Jane",
			email:    "jane@test.com",
			expected: "Hi Jane, we'll contact you at jane@test.com",
		},
		{
			name:     "multiple occurrences",
			text:     "{{name}} {{name}} {{email}} {{email}}",
			userName: "Bob",
			email:    "bob@mail.com",
			expected: "Bob Bob bob@mail.com bob@mail.com",
		},
		{
			name:     "no placeholders",
			text:     "Plain text without any placeholders",
			userName: "Alice",
			email:    "alice@example.com",
			expected: "Plain text without any placeholders",
		},
		{
			name:     "empty name",
			text:     "Hi {{name}}!",
			userName: "",
			email:    "test@example.com",
			expected: "Hi !",
		},
		{
			name:     "html content",
			text:     "<p>Hello {{name}},</p><p>Contact: {{email}}</p>",
			userName: "Test User",
			email:    "user@test.com",
			expected: "<p>Hello Test User,</p><p>Contact: user@test.com</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := worker.ReplaceTemplateVars(tt.text, tt.userName, tt.email)
			if result != tt.expected {
				t.Errorf("ReplaceTemplateVars() = %q, want %q", result, tt.expected)
			}
		})
	}
}
