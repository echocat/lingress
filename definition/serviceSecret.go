package definition

import (
	"errors"
	"fmt"
	"github.com/echocat/lingress/support"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

type ServiceSecret struct {
	*Definition

	client      kubernetes.Interface
	resyncAfter time.Duration
	namespace   string
}

func NewServiceSecrets(client kubernetes.Interface, resyncAfter time.Duration) (*ServiceSecret, error) {
	if definition, err := newDefinition("service secrets", nil); err != nil {
		return nil, err
	} else {
		return &ServiceSecret{
			Definition:  definition,
			client:      client,
			resyncAfter: resyncAfter,
		}, nil
	}
}

func (instance *ServiceSecret) SetNamespace(namespace string) {
	instance.namespace = namespace
}

func (instance *ServiceSecret) Init(stop support.Channel) error {
	namespace := instance.namespace
	if namespace == "" {
		return errors.New("no namespace for service provided")
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		instance.client,
		instance.resyncAfter,
		informers.WithNamespace(namespace),
	)
	instance.SetInformer(informerFactory.Core().V1().Secrets().Informer())

	return instance.Definition.Init(stop)
}

func (instance *ServiceSecret) Get(key string) (*v1.Service, error) {
	if item, exists, err := instance.informer.GetStore().GetByKey(key); err != nil {
		return nil, fmt.Errorf("cannot get secrets %s from cache: %v", key, err)
	} else if !exists {
		return nil, nil
	} else {
		return item.(*v1.Service), nil
	}
}
