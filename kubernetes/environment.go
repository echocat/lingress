package kubernetes

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"io/ioutil"
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

func NewEnvironment() (env *Environment, err error) {
	return &Environment{
		lock:          new(sync.Mutex),
		tokenFile:     ServiceTokenFile,
		rootCAFile:    ServiceRootCAFile,
		namespaceFile: ServiceNamespace,
	}, nil
}

func MustNewEnvironment() *Environment {
	result, err := NewEnvironment()
	support.Must(err)
	return result
}

type Environment struct {
	Kubeconfig KubeconfigPath
	Context    string
	Namespace  string

	lock    *sync.Mutex
	payload *environmentPayload

	tokenFile     string
	rootCAFile    string
	namespaceFile string
}

type environmentPayload struct {
	restConfig  *rest.Config
	contextName string
}

func (instance *Environment) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("kubeconfig", "Defines the location of the kubeconfig."+
		" If set to 'incluster' it will assume it is inside of the cluster."+
		" If set to 'mock' a mocked version will be created that works with every context together.").
		PlaceHolder("<kube config file>").
		Envar(support.FlagEnvName(appPrefix, "KUBECONFIG")).
		SetValue(&instance.Kubeconfig)
	fe.Flag("context", "Defines context which you try to access of the kubeconfig."+
		" This value is ignored if kbueconfig 'incluster' is used.").
		PlaceHolder("<context>").
		Short('c').
		Envar(support.FlagEnvName(appPrefix, "CONTEXT")).
		StringVar(&instance.Context)
	fe.Flag("namespace", "Defines namespace where it service is bound to/running in."+
		" This value is ignored if kbueconfig 'incluster' is used.").
		PlaceHolder("<namespace>").
		Short('n').
		Envar(support.FlagEnvName(appPrefix, "NAMESPACE")).
		StringVar(&instance.Namespace)
	return nil
}

func (instance *Environment) NewClient() (kubernetes.Interface, error) {
	if loaded, err := instance.get(); err != nil {
		return nil, err
	} else if loaded.restConfig == nil {
		// TODO! We should find a way to implement this, too
		return nil, fmt.Errorf("currently there is no support for mock of %v", reflect.TypeOf((*kubernetes.Interface)(nil)).Elem())
	} else {
		return kubernetes.NewForConfig(loaded.restConfig)
	}
}

func (instance *Environment) NewDynamicClient() (dynamic.Interface, error) {
	if loaded, err := instance.get(); err != nil {
		return nil, err
	} else if loaded.restConfig == nil {
		scheme := runtime.NewScheme()
		return dynamicFake.NewSimpleDynamicClient(scheme), nil
	} else {
		return dynamic.NewForConfig(loaded.restConfig)
	}
}

func (instance *Environment) ContextName() (string, error) {
	if loaded, err := instance.get(); err != nil {
		return "", err
	} else {
		return loaded.contextName, nil
	}
}

func (instance *Environment) get() (loaded *environmentPayload, err error) {
	if instance.payload == nil {
		instance.lock.Lock()
		defer instance.lock.Unlock()
		if instance.payload == nil {
			if payload, err := instance.load(instance.Kubeconfig, instance.Context); err != nil {
				return nil, err
			} else {
				instance.payload = payload
			}
		}
	}
	return instance.payload, nil
}

func (instance *Environment) load(path KubeconfigPath, targetContext string) (*environmentPayload, error) {
	if path.IsInCluster() {
		return instance.loadForInCluster()
	} else if path.IsMock() {
		return instance.loadMock(targetContext)
	} else {
		return instance.loadFromPath(path, targetContext)
	}
}

func (instance *Environment) loadForInCluster() (*environmentPayload, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, rest.ErrNotInCluster
	}

	ts := transport.NewCachedFileTokenSource(instance.tokenFile)

	if _, err := ts.Token(); err != nil {
		return nil, err
	}

	tlsClientConfig := rest.TLSClientConfig{}

	if _, err := certutil.NewPool(instance.rootCAFile); err != nil {
		return nil, fmt.Errorf("expected to load root CA config from %s, but got err: %v", instance.rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = instance.rootCAFile
	}

	if nsb, err := ioutil.ReadFile(instance.namespaceFile); err != nil {
		return nil, fmt.Errorf("expected to load namespace from %s, but got err: %v", instance.rootCAFile, err)
	} else {
		instance.Namespace = string(nsb)
	}

	return &environmentPayload{
		restConfig: &rest.Config{
			Host:            "https://" + net.JoinHostPort(host, port),
			TLSClientConfig: tlsClientConfig,
			WrapTransport:   transport.TokenSourceWrapTransport(ts),
		},
		contextName: KubeconfigInCluster,
	}, nil
}

func (instance *Environment) loadMock(targetContext string) (*environmentPayload, error) {
	if targetContext == "" {
		targetContext = "mock"
	}
	return &environmentPayload{
		restConfig:  nil,
		contextName: targetContext,
	}, nil
}

func (instance *Environment) loadFromPath(path KubeconfigPath, targetContext string) (*environmentPayload, error) {
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
