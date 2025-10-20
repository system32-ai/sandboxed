package main

import (
	"log"

	"github.com/altgen-ai/sandboxed/pkg/sdk"
)

func main() {

	sandbox, err := sdk.CreateSandbox("debug-generated-code", sdk.Python)
	if err != nil {
		log.Fatalf("failed to create sandbox: %v", err)
	}

	defer sandbox.Destroy()

	code := `python -c 'print("Hello, World!")'`

	output, err := sandbox.Run(code)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Output: %s", output.Result)

	code = `python --version`
	output, err = sandbox.Run(code)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Output: %s", output.Result)
}