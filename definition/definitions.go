package definition

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sync/atomic"
	"time"
)

type Definitions struct {
	Ingress *Ingress
	Service *Service
}

func New(client kubernetes.Interface, resyncAfter time.Duration) (*Definitions, error) {
	if ingress, err := NewIngress(client, resyncAfter); err != nil {
		return nil, fmt.Errorf("cannot create ingress definition store: %v", err)
	} else if service, err := NewService(client, resyncAfter); err != nil {
		return nil, fmt.Errorf("cannot create service definition store: %v", err)
	} else {
		return &Definitions{
			Ingress: ingress,
			Service: service,
		}, nil
	}
}

func (instance *Definitions) Init(stopCh chan struct{}) error {
	var pErr atomic.Value
	var done uint32

	instance.configureErrorHandler(&pErr, &done, instance.Ingress.Definition)
	instance.configureErrorHandler(&pErr, &done, instance.Service.Definition)

	go instance.Service.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, instance.Service.HasSynced) {
		return fmt.Errorf("initial service synchronization failed")
	} else {
		if plainErr := pErr.Load(); plainErr != nil {
			return plainErr.(error)
		}
	}

	go instance.Ingress.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, instance.Ingress.HasSynced) {
		return fmt.Errorf("initial ingress synchronization failed")
	} else {
		atomic.StoreUint32(&done, 1)
		if plainErr := pErr.Load(); plainErr != nil {
			return plainErr.(error)
		} else {
			return nil
		}
	}
}

func (instance *Definitions) configureErrorHandler(pErr *atomic.Value, done *uint32, target *Definition) {
	target.OnError = func(key string, event string, err error) {
		if atomic.LoadUint32(done) <= 0 {
			pErr.Store(err)
			target.logKey(event, key).
				WithError(err).
				Warn("cannot sync definition; will fail")
		} else {
			target.logKey(event, key).
				WithError(err).
				Error("cannot sync definition")
		}
	}
}

func (instance *Definitions) HasSynced() bool {
	return instance.Ingress.HasSynced() &&
		instance.Service.HasSynced()
}
