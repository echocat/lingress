package definition

import (
	"errors"
	"fmt"
	"github.com/echocat/lingress/support"
	log "github.com/echocat/slf4g"
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

func NewServiceSecrets(client kubernetes.Interface, resyncAfter time.Duration, logger log.Logger) (*ServiceSecret, error) {
	if definition, err := newDefinition("service-secrets", nil, logger); err != nil {
		return nil, err
	} else {
		return &ServiceSecret{
			Definition:  definition,
			client:      client,
			resyncAfter: resyncAfter,
		}, nil
	}
}

func (this *ServiceSecret) SetNamespace(namespace string) {
	this.namespace = namespace
}

func (this *ServiceSecret) Init(stop support.Channel) error {
	namespace := this.namespace
	if namespace == "" {
		return errors.New("no namespace for service provided")
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		this.client,
		this.resyncAfter,
		informers.WithNamespace(namespace),
	)
	this.SetInformer(informerFactory.Core().V1().Secrets().Informer())

	return this.Definition.Init(stop)
}

func (this *ServiceSecret) Get(key string) (*v1.Service, error) {
	if item, exists, err := this.informer.GetStore().GetByKey(key); err != nil {
		return nil, fmt.Errorf("cannot get secrets %s from cache: %v", key, err)
	} else if !exists {
		return nil, nil
	} else {
		return item.(*v1.Service), nil
	}
}
