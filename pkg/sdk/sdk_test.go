package sdk_test

import (
	"os"
	"testing"

	"github.com/system32-ai/sandboxed/pkg/sdk"
)

func TestSimpleCodeRun(t *testing.T) {

	sandbox, err := sdk.CreateSandbox("test", sdk.Python)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	code := `python -c 'print("Hello, World!")'`

	output, err := sandbox.Run(code)
	if err != nil {
		t.Fatalf("failed to run code: %v", err)
	}

	t.Logf("Output: %s", output.Result)

	code = `python --version`
	output, err = sandbox.Run(code)
	if err != nil {
		t.Fatalf("failed to run code: %v", err)
	}

	t.Logf("Output: %s", output.Result)

	sandbox.Destroy()
}

func TestCreateSandboxAuto_NoAPIKey(t *testing.T) {
	// Ensure no API key is set
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	pythonCode := `
import sys
print("Hello from Python!")
print(f"Python version: {sys.version}")
`

	_, err := sdk.CreateSandboxAuto("test-auto", pythonCode)
	if err == nil {
		t.Fatal("expected error when no OPENAI_API_KEY is set")
	}

	expectedError := "OPENAI_API_KEY environment variable is required"
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDetectLanguage_NoAPIKey(t *testing.T) {
	// Ensure no API key is set
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	pythonCode := `print("Hello, World!")`

	_, err := sdk.DetectLanguage(pythonCode)
	if err == nil {
		t.Fatal("expected error when no OPENAI_API_KEY is set")
	}

	expectedError := "OPENAI_API_KEY environment variable is required"
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}
