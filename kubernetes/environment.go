package kubernetes

import (
	"fmt"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
	certutil "k8s.io/client-go/util/cert"
	"net"
	"os"
	"reflect"
	"sync"
)

const (
	ServiceTokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	ServiceRootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	ServiceNamespace  = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

func NewEnvironment(s *settings.Settings) (env *Environment, err error) {
	return &Environment{
		settings:      s,
		lock:          new(sync.Mutex),
		tokenFile:     ServiceTokenFile,
		rootCAFile:    ServiceRootCAFile,
		namespaceFile: ServiceNamespace,
	}, nil
}

func MustNewEnvironment(s *settings.Settings) *Environment {
	result, err := NewEnvironment(s)
	support.Must(err)
	return result
}

type Environment struct {
	settings *settings.Settings
	lock     *sync.Mutex
	payload  *environmentPayload

	tokenFile     string
	rootCAFile    string
	namespaceFile string
}

type environmentPayload struct {
	restConfig  *rest.Config
	contextName string
}

func (this *Environment) NewClient() (kubernetes.Interface, error) {
	if loaded, err := this.get(); err != nil {
		return nil, err
	} else if loaded.restConfig == nil {
		// TODO! We should find a way to implement this, too
		return nil, fmt.Errorf("currently there is no support for mock of %v", reflect.TypeOf((*kubernetes.Interface)(nil)).Elem())
	} else {
		return kubernetes.NewForConfig(loaded.restConfig)
	}
}

func (this *Environment) NewDynamicClient() (dynamic.Interface, error) {
	if loaded, err := this.get(); err != nil {
		return nil, err
	} else if loaded.restConfig == nil {
		scheme := runtime.NewScheme()
		return dynamicFake.NewSimpleDynamicClient(scheme), nil
	} else {
		return dynamic.NewForConfig(loaded.restConfig)
	}
}

func (this *Environment) ContextName() (string, error) {
	if loaded, err := this.get(); err != nil {
		return "", err
	} else {
		return loaded.contextName, nil
	}
}

func (this *Environment) get() (loaded *environmentPayload, err error) {
	if this.payload == nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		if this.payload == nil {
			if payload, err := this.load(this.settings.Kubernetes.Config, this.settings.Kubernetes.Context); err != nil {
				return nil, err
			} else {
				this.payload = payload
			}
		}
	}
	return this.payload, nil
}

func (this *Environment) load(path settings.KubeconfigPath, targetContext string) (*environmentPayload, error) {
	if path.IsInCluster() {
		return this.loadForInCluster()
	} else if path.IsMock() {
		return this.loadMock(targetContext)
	} else {
		return this.loadFromPath(path, targetContext)
	}
}

func (this *Environment) loadForInCluster() (*environmentPayload, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, rest.ErrNotInCluster
	}

	ts := transport.NewCachedFileTokenSource(this.tokenFile)

	if _, err := ts.Token(); err != nil {
		return nil, err
	}

	tlsClientConfig := rest.TLSClientConfig{}

	if _, err := certutil.NewPool(this.rootCAFile); err != nil {
		return nil, fmt.Errorf("expected to load root CA config from %s, but got err: %v", this.rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = this.rootCAFile
	}

	if nsb, err := os.ReadFile(this.namespaceFile); err != nil {
		return nil, fmt.Errorf("expected to load namespace from %s, but got err: %v", this.rootCAFile, err)
	} else {
		this.settings.Kubernetes.Namespace = string(nsb)
	}

	return &environmentPayload{
		restConfig: &rest.Config{
			Host:            "https://" + net.JoinHostPort(host, port),
			TLSClientConfig: tlsClientConfig,
			WrapTransport:   transport.TokenSourceWrapTransport(ts),
		},
		contextName: settings.KubeconfigInCluster,
	}, nil
}

func (this *Environment) loadMock(targetContext string) (*environmentPayload, error) {
	if targetContext == "" {
		targetContext = "mock"
	}
	return &environmentPayload{
		restConfig:  nil,
		contextName: targetContext,
	}, nil
}

func (this *Environment) loadFromPath(path settings.KubeconfigPath, targetContext string) (*environmentPayload, error) {
	loadedConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		path.ToConfigLoader(targetContext),
		&clientcmd.ConfigOverrides{
			CurrentContext: targetContext,
		},
	)

	if rc, err := loadedConfig.RawConfig(); err != nil {
		return nil, err
	} else if rc.CurrentContext == "" {
		return nil, clientcmd.ErrNoContext
	} else if restConfig, err := loadedConfig.ClientConfig(); err != nil {
		return nil, err
	} else {
		return &environmentPayload{
			restConfig:  restConfig,
			contextName: rc.CurrentContext,
		}, nil
	}
}
