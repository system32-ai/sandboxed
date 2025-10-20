package sdk_test

import (
	"testing"

	"github.com/altgen.ai/sandboxed/pkg/sdk"
)

func TestCreateSandbox(t *testing.T) {
	sandboxed := sdk.NewSandboxed()

	tests := []struct {
		lang          string
		expectedImage string
		expectError   bool
	}{
		{"python", "python:3.9-slim", false},
		{"javascript", "node:14-alpine", false},
		{"go", "golang:1.16-alpine", false},
		{"ruby", "ruby:2.7-alpine", false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			container, err := sandboxed.CreateSandbox("test",tt.lang)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for language %s, got none", tt.lang)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error for language %s, got %v", tt.lang, err)
				}

				container.Run("echo test")
			}
		})
	}
}

func TestSimpleCodeRun(t *testing.T) {
	sandboxed := sdk.NewSandboxed()

	container, err := sandboxed.CreateSandbox("test", "python")
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	code := []string{
		`python -c "print('Hello, World!')"`,
	}

	output, err := container.Code(code)
	if err != nil {
		t.Fatalf("failed to run code: %v", err)
	}

	expectedOutput := "Hello, World!\n"
	if output.Result != expectedOutput {
		t.Errorf("expected output %q, got %q", expectedOutput, output.Result)
	}
}