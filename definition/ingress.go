package definition

import (
	"fmt"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Ingress struct {
	*Definition
}

func NewIngress(client kubernetes.Interface, resyncAfter time.Duration) (*Ingress, error) {
	informerFactory := informers.NewSharedInformerFactory(client, resyncAfter)
	informer := informerFactory.Extensions().V1beta1().Ingresses().Informer()
	if definition, err := newDefinition("ingress", informer); err != nil {
		return nil, err
	} else {
		return &Ingress{
			Definition: definition,
		}, nil
	}
}

func (instance *Ingress) Get(key string) (*v1beta1.Ingress, error) {
	if item, exists, err := instance.informer.GetStore().GetByKey(key); err != nil {
		return nil, fmt.Errorf("cannot get ingress %s from cache: %v", key, err)
	} else if !exists {
		return nil, nil
	} else {
		return item.(*v1beta1.Ingress), nil
	}
}
