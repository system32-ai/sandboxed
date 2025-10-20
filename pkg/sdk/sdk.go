package sdk

import (
	"errors"
	"time"

	"github.com/altgen.ai/sandboxed/pkg/k8sclient"
	"github.com/altgen.ai/sandboxed/pkg/k8sclient/templates"
)

type SandboxOption struct {
	Name string
	Value interface{}
}

type Sandboxed interface {
	CreateSandbox(name string, lang string, opts ...SandboxOption) (*LanguageContainer, error)
	Run(name, image string, code []string, opts ...SandboxOption) (*Output, error)
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
}

func (s *sandboxedImpl) CreateSandbox(name, lang string, opts ...SandboxOption) (*LanguageContainer, error) {
	image, err := templates.LanguageLookup(lang)
	if err != nil {
		return nil, err
	}

	return &LanguageContainer{
		name:    name,
		language: lang,
		image:    image,
		impl:    s,
		opts:    opts,
	}, nil
}

func (s *sandboxedImpl) Run(name, image string, code []string, opts ...SandboxOption) (*Output, error) {

	var mapOptions = make(map[string]interface{})
	for _, opt := range opts {
		mapOptions[opt.Name] = opt.Value
	}

	var namespace string
	var podName = "sandboxed-" + name

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

	pod.Image = image
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

	var cmds []string
	
	for _, line := range code {
		cmds = append(cmds, "sh -c '"+line+"'")
	}

	o, err := client.ExecCommand(podName, pod.Namespace, cmds)
	if err != nil {
		return nil, err
	}

	defer client.DeletePod(podName, pod.Namespace)

	return &Output{
		Result: o,
		Error:    "",
		ExitCode: 0,
	}, nil
}

