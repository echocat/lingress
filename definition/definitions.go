package definition

import (
	"fmt"
	"github.com/echocat/lingress/support"
	log "github.com/echocat/slf4g"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Definitions struct {
	ServiceSecrets *ServiceSecret
	Ingress        *Ingress
	Service        *Service
}

func New(client kubernetes.Interface, resyncAfter time.Duration, logger log.Logger) (*Definitions, error) {
	if serviceSecrets, err := NewServiceSecrets(client, resyncAfter, logger); err != nil {
		return nil, fmt.Errorf("cannot create service secrets definition store: %v", err)
	} else if ingress, err := NewIngress(client, resyncAfter, logger); err != nil {
		return nil, fmt.Errorf("cannot create ingress definition store: %v", err)
	} else if service, err := NewService(client, resyncAfter, logger); err != nil {
		return nil, fmt.Errorf("cannot create service definition store: %v", err)
	} else {
		return &Definitions{
			ServiceSecrets: serviceSecrets,
			Ingress:        ingress,
			Service:        service,
		}, nil
	}
}

func (instance *Definitions) SetNamespace(namespace string) {
	instance.ServiceSecrets.SetNamespace(namespace)
}

func (instance *Definitions) Init(stop support.Channel) error {
	if err := instance.ServiceSecrets.Init(stop); err != nil {
		return err
	}

	if err := instance.Service.Init(stop); err != nil {
		return err
	}

	if err := instance.Ingress.Init(stop); err != nil {
		return err
	}

	return nil
}

func (instance *Definitions) HasSynced() bool {
	return instance.Ingress.HasSynced() &&
		instance.Service.HasSynced() &&
		instance.ServiceSecrets.HasSynced()
}
