package sdk

type LanguageContainer struct {
	name     string
	language string
	image    string
	impl    *sandboxedImpl
	opts    []SandboxOption
	code	string
}

type Output struct {
	Result string
	Error  string
	ExitCode int
}