package sdk

import (
	"errors"
	"time"

	"github.com/altgen-ai/sandboxed/pkg/k8sclient"
	"github.com/altgen-ai/sandboxed/pkg/k8sclient/templates"
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

type SandboxOption struct {
	Name string
	Value interface{}
}

type Sandboxed interface {
	Run(code string) (*Output, error)
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

type sandboxedImpl struct{
	driver string
	id string
	lc *LanguageContainer
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
		name:    name,
		language: string(lang),
		image:    image,
		impl:    s,
		opts:    opts,
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
		return nil	, errors.New("unsupported driver: " + s.driver)
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

	s.id  = podName

	return s, nil
}


func NewInstance(id string, opts ...SandboxOption) (Sandboxed, error) {
	
	s := &sandboxedImpl{
		driver: "kubernetes",
	}


	lcVal := &LanguageContainer{
		// name:    id,
		// language: lang,
		// image:    image,
		impl:    s,
		opts:    opts,
	}

	s.lc = lcVal
	s.id = id

	return s, nil
}

func (s *sandboxedImpl) Run( code string) (*Output, error) {

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
		return nil	, errors.New("unsupported driver: " + s.driver)
	}

	o, err := client.ExecCommand(podName, namespace, []string{"sh", "-c", code})
	if err != nil {
		return nil, err
	}

	return &Output{
		Result: o,
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