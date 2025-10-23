package sdk_test

import (
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
