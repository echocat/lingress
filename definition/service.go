package definition

import (
	"fmt"
	log "github.com/echocat/slf4g"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Service struct {
	*Definition
}

func NewService(client kubernetes.Interface, resyncAfter time.Duration, logger log.Logger) (*Service, error) {
	informerFactory := informers.NewSharedInformerFactory(client, resyncAfter)
	informer := informerFactory.Core().V1().Services().Informer()
	if definition, err := newDefinition("service", informer, logger); err != nil {
		return nil, err
	} else {
		return &Service{
			Definition: definition,
		}, nil
	}
}

func (instance *Service) Get(key string) (*v1.Service, error) {
	if item, exists, err := instance.informer.GetStore().GetByKey(key); err != nil {
		return nil, fmt.Errorf("cannot get service %s from cache: %v", key, err)
	} else if !exists {
		return nil, nil
	} else {
		return item.(*v1.Service), nil
	}
}
