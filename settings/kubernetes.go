package settings

import (
	"dario.cat/mergo"
	"errors"
	"fmt"
	"github.com/echocat/lingress/support"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/user"
	"path/filepath"
)

const (
	KubeconfigInCluster = "incluster"
	KubeconfigMock      = "mock"
	EnvVarKubeconfig    = "KUBE_CONFIG"
)

func NewKubernetes() (Kubernetes, error) {
	return Kubernetes{}, nil
}

type Kubernetes struct {
	Config    KubeconfigPath `json:"config,omitempty" yaml:"config,omitempty"`
	Context   string         `json:"context,omitempty" yaml:"context,omitempty"`
	Namespace string         `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

func (this *Kubernetes) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("kubernetes.config", fmt.Sprintf("Defines the location of the kubeconfig."+
		" If set to '%s' it will assume it is inside of the cluster."+
		" If set to 'mock' a mocked version will be created that works with every context together.", KubeconfigInCluster)).
		PlaceHolder("<kube config file>").
		Envar(support.FlagEnvName(appPrefix, "KUBECONFIG")).
		SetValue(&this.Config)
	fe.Flag("kubernetes.context", fmt.Sprintf("Defines context which you try to access of the kubeconfig."+
		" This value is ignored if kubeconfig '%s' is used.", KubeconfigInCluster)).
		PlaceHolder("<context>").
		Short('c').
		Envar(support.FlagEnvName(appPrefix, "KUBERNETES_CONTEXT")).
		StringVar(&this.Context)
	fe.Flag("kubernetes.namespace", fmt.Sprintf("Defines namespace where it service is bound to/running in."+
		" This value is ignored if kubeconfig '%s' is used.", KubeconfigInCluster)).
		PlaceHolder("<namespace>").
		Short('n').
		Envar(support.FlagEnvName(appPrefix, "KUBERNETES_NAMESPACE")).
		StringVar(&this.Namespace)
}

type KubeconfigPath struct {
	Value string

	ResolveDefaultLocation func() string
}

func (this KubeconfigPath) String() string {
	if this.Value == "" {
		return this.resolveDefaultLocation()
	}
	return this.Value
}

func (this KubeconfigPath) IsEmpty() bool {
	return this.Value == ""
}

func (this *KubeconfigPath) Set(plain string) error {
	if plain == "" || plain == KubeconfigInCluster || plain == KubeconfigMock {
		this.Value = plain
		return nil
	} else if fi, err := os.Stat(plain); err != nil {
		return &os.PathError{Op: "set_kubeconfig", Path: plain, Err: err}
	} else if !fi.IsDir() {
		return &os.PathError{Op: "set_kubeconfig", Path: plain, Err: errors.New("no file")}
	} else {
		this.Value = plain
		return nil
	}
}

func (this KubeconfigPath) IsInCluster() bool {
	return this.Value == KubeconfigInCluster
}

func (this KubeconfigPath) IsMock() bool {
	return this.Value == KubeconfigMock
}

func (this KubeconfigPath) resolveDefaultLocation() string {
	if rd := this.ResolveDefaultLocation; rd != nil {
		return rd()
	} else if u, err := user.Current(); err != nil {
		return filepath.Join("~", ".kube", "config")
	} else {
		return filepath.Join(u.HomeDir, ".kube", "config")
	}
}

func (this KubeconfigPath) ToConfigLoader(targetContext string) KubeconfigLoader {
	return KubeconfigLoader{
		Kubeconfig:        this.Value,
		DefaultKubeconfig: this.resolveDefaultLocation(),
		TargetContext:     targetContext,
	}
}

type KubeconfigLoader struct {
	clientcmd.ClientConfigLoader
	Kubeconfig        string
	DefaultKubeconfig string
	TargetContext     string
}

func (this KubeconfigLoader) Load() (*clientcmdapi.Config, error) {
	config := clientcmdapi.NewConfig()
	config.CurrentContext = this.TargetContext
	atLeastOneConfigFound := false

	if plainFromEnv, ok := os.LookupEnv(EnvVarKubeconfig); ok {
		if fromEnv, err := clientcmd.Load([]byte(plainFromEnv)); err != nil {
			return nil, err
		} else if err := mergo.Merge(config, fromEnv); err != nil {
			return nil, err
		} else {
			atLeastOneConfigFound = true
		}
	}

	targetKubeConfigPath := this.Kubeconfig
	if targetKubeConfigPath != "" {
		if _, err := os.Stat(targetKubeConfigPath); err != nil {
			return nil, err
		}
	} else {
		targetKubeConfigPath = this.DefaultKubeconfig
	}

	if targetKubeConfigPath != "" {
		if fromFile, err := clientcmd.LoadFromFile(targetKubeConfigPath); os.IsNotExist(err) {
			// Ignore and continue
		} else if err != nil {
			return nil, err
		} else if err := mergo.Merge(config, fromFile); err != nil {
			return nil, err
		} else {
			atLeastOneConfigFound = true
		}
	}

	if !atLeastOneConfigFound {
		return nil, fmt.Errorf("there is neither argument --kubernetes.config nor environment variable %s provided nor does %s exist", EnvVarKubeconfig, this.DefaultKubeconfig)
	}

	return config, nil
}

func (this KubeconfigLoader) IsDefaultConfig(*restclient.Config) bool {
	return false
}
