package definition

import (
	"fmt"
	log "github.com/echocat/slf4g"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Ingress struct {
	*Definition
}

func NewIngress(client kubernetes.Interface, resyncAfter time.Duration, logger log.Logger) (*Ingress, error) {
	informerFactory := informers.NewSharedInformerFactory(client, resyncAfter)
	informer := informerFactory.Networking().V1().Ingresses().Informer()
	if definition, err := newDefinition("ingress", informer, logger); err != nil {
		return nil, err
	} else {
		return &Ingress{
			Definition: definition,
		}, nil
	}
}

func (this *Ingress) Get(key string) (*networkingv1.Ingress, error) {
	if item, exists, err := this.informer.GetStore().GetByKey(key); err != nil {
		return nil, fmt.Errorf("cannot get ingress %s from cache: %v", key, err)
	} else if !exists {
		return nil, nil
	} else {
		return item.(*networkingv1.Ingress), nil
	}
}
