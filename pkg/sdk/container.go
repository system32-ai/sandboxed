package sdk

type LanguageContainer struct {
	name     string
	language string
	image    string
	impl    *sandboxedImpl
	opts    []SandboxOption
	code	[]string
}

type Output struct {
	Result string
	Error  string
	ExitCode int
}

func (lc *LanguageContainer) Run(cmd string) (*Output, error) {
	return lc.impl.Run(lc.name, lc.image, lc.code, lc.opts...)
}

func (lc *LanguageContainer) Code(code []string) (*Output, error) {
	output, err := lc.impl.Run(lc.name, lc.image, code, lc.opts...)
	if err != nil {
		return nil, err
	}

	return output, nil
}