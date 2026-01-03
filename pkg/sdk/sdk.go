package sdk

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/system32-ai/sandboxed/pkg/k8sclient"
	"github.com/system32-ai/sandboxed/pkg/k8sclient/templates"
)

type Language string

const (
	Python Language = "python"
	Go     Language = "go"
	Node   Language = "node"
	Java   Language = "java"
	Ruby   Language = "ruby"
	PHP    Language = "php"
	Rust   Language = "rust"
)

func (l Language) GetExecScript() string {
	switch l {
	case Python:
		return "python3 /tmp/exec_script.sh"
	case Go:
		return "go run /tmp/exec_script.go"
	case Node:
		return "node /tmp/exec_script.js"
	case Java:
		return "javac /tmp/ExecScript.java && java -cp /tmp ExecScript"
	case Ruby:
		return "ruby /tmp/exec_script.rb"
	case PHP:
		return "php /tmp/exec_script.php"
	case Rust:
		return "rustc /tmp/exec_script.rs -o /tmp/exec_script && /tmp/exec_script"
	default:
		return ""
	}
}

func (l Language) DockerImage() (string, error) {
	return templates.LanguageLookup(string(l))
}

func (l Language) IsValid() bool {
	_, err := templates.LanguageLookup(string(l))
	return err == nil
}

func ToLanguage(lang string) (Language, error) {
	if _, err := templates.LanguageLookup(lang); err != nil {
		return "", err
	}
	return Language(lang), nil
}

// DetectLanguage uses GPT to analyze the code string and determine the programming language
func DetectLanguage(code string) (Language, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY environment variable is required for automatic language detection")
	}

	client := openai.NewClient(apiKey)

	prompt := `Analyze the following code and determine the primary programming language it is written in. 
Respond with only the language name in lowercase (one of: python, go, node, java, ruby, php, rust).

Code:
` + code

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   10,
			Temperature: 0,
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from GPT API")
	}

	detectedLang := strings.TrimSpace(strings.ToLower(resp.Choices[0].Message.Content))

	// Map the response to our Language constants
	switch detectedLang {
	case "python":
		return Python, nil
	case "go":
		return Go, nil
	case "node", "javascript", "js":
		return Node, nil
	case "java":
		return Java, nil
	case "ruby":
		return Ruby, nil
	case "php":
		return PHP, nil
	case "rust":
		return Rust, nil
	default:
		return "", errors.New("unable to detect supported programming language from code")
	}
}

type SandboxOption struct {
	Name  string
	Value interface{}
}

type Sandboxed interface {
	Run(code string) (*Output, error)
	Exec(commands string) (*Output, error)
	Destroy() error
}

func NewSandboxed() Sandboxed {
	return &sandboxedImpl{
		driver: "kubernetes",
	}
}

func NewSandboxForDocker() Sandboxed {
	return &sandboxedImpl{
		driver: "docker",
	}
}

type sandboxedImpl struct {
	driver string
	id     string
	lc     *LanguageContainer
}

func CreateSandbox(name string, lang Language, opts ...SandboxOption) (Sandboxed, error) {

	s := &sandboxedImpl{
		driver: "kubernetes",
	}

	var client *k8sclient.Client
	var err error

	image, err := templates.LanguageLookup(string(lang))
	if err != nil {
		return nil, err
	}

	lcVal := &LanguageContainer{
		name:     name,
		language: string(lang),
		image:    image,
		impl:     s,
		opts:     opts,
	}

	s.lc = lcVal

	var mapOptions = make(map[string]interface{})
	for _, opt := range s.lc.opts {
		mapOptions[opt.Name] = opt.Value
	}

	var namespace string
	var podName = "sandboxed-" + s.lc.name

	ns, ok := mapOptions["namespace"].(string)
	if ok {
		namespace = ns
	} else {
		namespace = "default"
	}

	if s.driver == "kubernetes" {
		client, err = k8sclient.NewClient(namespace)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unsupported driver: " + s.driver)
	}

	var pod k8sclient.PodSpec

	pod.Labels, ok = mapOptions["labels"].(map[string]string)
	if !ok {
		pod.Labels = make(map[string]string)
	} else {
		for k, v := range pod.Labels {
			pod.Labels[k] = v
		}
		pod.Labels["created-by"] = "sandboxed-sdk"
	}

	pod.Image = s.lc.image
	pod.Name = podName
	pod.Namespace = namespace
	pod.Command = []string{"sh", "-c", "tail -f /dev/null"}

	_, err = client.CreatePod(pod)
	if err != nil {
		return nil, err
	}

	if err := client.WaitForPodReady(podName, pod.Namespace, 120*time.Second); err != nil {
		return nil, err
	}

	s.id = podName

	return s, nil
}

// CreateSandboxAuto creates a sandbox with automatic language detection from the provided code
func CreateSandboxAuto(name string, code string, opts ...SandboxOption) (Sandboxed, error) {
	lang, err := DetectLanguage(code)
	if err != nil {
		return nil, err
	}

	return CreateSandbox(name, lang, opts...)
}

func NewInstance(id string, opts ...SandboxOption) (Sandboxed, error) {

	s := &sandboxedImpl{
		driver: "kubernetes",
	}

	lcVal := &LanguageContainer{
		// name:    id,
		// language: lang,
		// image:    image,
		impl: s,
		opts: opts,
	}

	s.lc = lcVal
	s.id = id

	return s, nil
}

func (s *sandboxedImpl) Run(code string) (*Output, error) {

	var client *k8sclient.Client
	var err error

	var mapOptions = make(map[string]interface{})
	for _, opt := range s.lc.opts {
		mapOptions[opt.Name] = opt.Value
	}

	var namespace string
	var podName = s.id

	ns, ok := mapOptions["namespace"].(string)
	if ok {
		namespace = ns
	} else {
		namespace = "default"
	}

	if s.driver == "kubernetes" {
		client, err = k8sclient.NewClient(namespace)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unsupported driver: " + s.driver)
	}

	o, err := client.ExecCommand(podName, namespace, []string{"sh", "-c", code})
	if err != nil {
		return nil, err
	}

	return &Output{
		Result:   o,
		Error:    "",
		ExitCode: 0,
	}, nil
}

func (s *sandboxedImpl) Exec(commands string) (*Output, error) {

	var client *k8sclient.Client
	var err error

	var mapOptions = make(map[string]interface{})
	for _, opt := range s.lc.opts {
		mapOptions[opt.Name] = opt.Value
	}

	var namespace string
	var podName = s.id
	ns, ok := mapOptions["namespace"].(string)
	if ok {
		namespace = ns
	} else {
		namespace = "default"
	}

	if s.driver == "kubernetes" {
		client, err = k8sclient.NewClient(namespace)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unsupported driver: " + s.driver)
	}

	// Write commands to a temporary file and execute it
	filename := "/tmp/exec_script.sh"
	writeCmd := []string{"sh", "-c", "cat > " + filename + " << 'EOF'\n" + commands + "\nEOF"}

	// First, write the commands to a file
	_, err = client.ExecCommand(podName, namespace, writeCmd)
	if err != nil {
		return nil, err
	}

	// Make the file executable
	chmodCmd := []string{"sh", "-c", "chmod +x " + filename}
	_, err = client.ExecCommand(podName, namespace, chmodCmd)
	if err != nil {
		return nil, err
	}

	lt, err := ToLanguage(s.lc.language)
	if err != nil {
		return nil, err
	}

	// Execute the file
	o, err := client.ExecCommand(podName, namespace, []string{"sh", "-c", lt.GetExecScript()})
	if err != nil {
		return nil, err
	}

	return &Output{
		Result:   o,
		Error:    "",
		ExitCode: 0,
	}, nil
}

func (s *sandboxedImpl) Destroy() error {

	var mapOptions = make(map[string]interface{})
	for _, opt := range s.lc.opts {
		mapOptions[opt.Name] = opt.Value
	}

	var namespace string

	ns, ok := mapOptions["namespace"].(string)
	if ok {
		namespace = ns
	} else {
		namespace = "default"
	}

	var client *k8sclient.Client
	var err error

	if s.driver == "kubernetes" {
		client, err = k8sclient.NewClient(namespace)
		if err != nil {
			return err
		}
	} else {
		return errors.New("unsupported driver: " + s.driver)
	}

	podName := "sandboxed-" + s.lc.name

	return client.ForceDeletePod(podName, namespace)
}
