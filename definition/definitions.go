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

func (this *Definitions) SetNamespace(namespace string) {
	this.ServiceSecrets.SetNamespace(namespace)
}

func (this *Definitions) Init(stop support.Channel) error {
	if err := this.ServiceSecrets.Init(stop); err != nil {
		return err
	}

	if err := this.Service.Init(stop); err != nil {
		return err
	}

	if err := this.Ingress.Init(stop); err != nil {
		return err
	}

	return nil
}

func (this *Definitions) HasSynced() bool {
	return this.Ingress.HasSynced() &&
		this.Service.HasSynced() &&
		this.ServiceSecrets.HasSynced()
}
