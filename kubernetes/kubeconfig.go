package kubernetes

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
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

var (
	ErrNoFile = errors.New("no file")
)

type KubeconfigPath struct {
	value string

	ResolveDefaultLocation func() string
}

func (instance KubeconfigPath) String() string {
	if instance.value == "" {
		return instance.resolveDefaultLocation()
	}
	return instance.value
}

func (instance KubeconfigPath) IsEmpty() bool {
	return instance.value == ""
}

func (instance *KubeconfigPath) Set(plain string) error {
	if plain == "" || plain == KubeconfigInCluster || plain == KubeconfigMock {
		instance.value = plain
		return nil
	} else if fi, err := os.Stat(plain); err != nil {
		return &os.PathError{Op: "set_kubeconfig", Path: plain, Err: err}
	} else if !fi.IsDir() {
		return &os.PathError{Op: "set_kubeconfig", Path: plain, Err: ErrNoFile}
	} else {
		instance.value = plain
		return nil
	}
}

func (instance KubeconfigPath) IsInCluster() bool {
	return instance.value == KubeconfigInCluster
}

func (instance KubeconfigPath) IsMock() bool {
	return instance.value == KubeconfigMock
}

func (instance KubeconfigPath) resolveDefaultLocation() string {
	if rd := instance.ResolveDefaultLocation; rd != nil {
		return rd()
	} else if u, err := user.Current(); err != nil {
		return filepath.Join("~", ".kube", "config")
	} else {
		return filepath.Join(u.HomeDir, ".kube", "config")
	}
}

func (instance KubeconfigPath) ToConfigLoader(targetContext string) ConfigLoader {
	return ConfigLoader{
		Kubeconfig:        instance.value,
		DefaultKubeconfig: instance.resolveDefaultLocation(),
		TargetContext:     targetContext,
	}
}

type ConfigLoader struct {
	clientcmd.ClientConfigLoader
	Kubeconfig        string
	DefaultKubeconfig string
	TargetContext     string
}

func (instance ConfigLoader) Load() (*clientcmdapi.Config, error) {
	config := clientcmdapi.NewConfig()
	config.CurrentContext = instance.TargetContext
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

	targetKubeConfigPath := instance.Kubeconfig
	if targetKubeConfigPath != "" {
		if _, err := os.Stat(targetKubeConfigPath); err != nil {
			return nil, err
		}
	} else {
		targetKubeConfigPath = instance.DefaultKubeconfig
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
		return nil, fmt.Errorf("there is neither argument --kubeconfig nor environment variable %s provided nor does %s exist", EnvVarKubeconfig, instance.DefaultKubeconfig)
	}

	return config, nil
}

func (instance ConfigLoader) IsDefaultConfig(*restclient.Config) bool {
	return false
}
