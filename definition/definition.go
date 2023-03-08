package definition

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"sync/atomic"
)

type Definition struct {
	MaxTries int

	typeDescription string
	informer        cache.SharedInformer
	lastError       atomic.Value

	OnElementAdded   OnElementChangedFunc
	OnElementUpdated OnElementUpdatedFunc
	OnElementRemoved OnElementRemovedFunc
	OnError          OnErrorFunc
}

type OnElementChangedFunc func(key string, new metav1.Object) error
type OnElementUpdatedFunc func(key string, old, new metav1.Object) error
type OnElementRemovedFunc func(key string) error
type OnErrorFunc func(key string, event string, err error)

func newDefinition(typeDescription string, informer cache.SharedInformer) (*Definition, error) {
	return &Definition{
		typeDescription: typeDescription,
		informer:        informer,
	}, nil
}

func (instance *Definition) SetInformer(informer cache.SharedInformer) {
	instance.informer = informer
}

func (instance *Definition) Init(stop support.Channel) error {
	if instance.informer == nil {
		panic(fmt.Sprintf("definition %s has no informer", instance.typeDescription))
	}

	_, err := instance.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    instance.onClusterElementAdded,
		UpdateFunc: instance.onClusterElementUpdated,
		DeleteFunc: instance.onClusterElementRemoved,
	})
	if err != nil {
		return fmt.Errorf("initial %s synchronization failed: %w", instance.typeDescription, err)
	}

	go instance.Run(stop)

	if !cache.WaitForCacheSync(support.ToChan(stop), instance.HasSynced) {
		stop.Broadcast()
		return fmt.Errorf("initial %s synchronization failed", instance.typeDescription)
	}
	if err := instance.lastError.Load(); err != nil {
		stop.Broadcast()
		return fmt.Errorf("initial %s synchronization failed: %w", instance.typeDescription, err.(error))
	}
	return nil
}

func (instance *Definition) HasSynced() bool {
	return instance.informer.HasSynced()
}

func (instance *Definition) Run(stop support.Channel) {
	l := instance.log()
	defer runtime.HandleCrash()
	l.Info("definition store started")

	go instance.informer.Run(support.ToChan(stop))
	stop.Wait()

	l.Info("definition store stopped")
}

func (instance *Definition) onClusterElementAdded(new interface{}) {
	l := instance.logEvent("elementAdded")

	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(new).String()).
			Error("cannot determine key of an object of type")
	}

	if instance.OnElementAdded == nil {
		return
	}

	l = l.With("key", key)
	if err := instance.OnElementAdded(key, new.(metav1.Object)); err != nil {
		instance.lastError.Store(err)
		if instance.OnError != nil {
			instance.OnError(key, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("cannot handle element")
		}
	} else {
		l.Debug("element added")
	}
}

func (instance *Definition) onClusterElementUpdated(old interface{}, new interface{}) {
	l := instance.logEvent("elementUpdated")

	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(new).String()).
			Error("cannot determine key of an object of type")
	}

	if instance.OnElementUpdated == nil {
		return
	}

	l = l.With("key", key)
	if err := instance.OnElementUpdated(key, old.(metav1.Object), new.(metav1.Object)); err != nil {
		instance.lastError.Store(err)
		if instance.OnError != nil {
			instance.OnError(key, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("cannot handle element")
		}
	} else {
		l.Debug("element updated")
	}
}

func (instance *Definition) onClusterElementRemoved(old interface{}) {
	l := instance.logEvent("elementRemoved")

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(old)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(old).String()).
			Error("cannot determine key of an object of type")
	}

	if instance.OnElementRemoved == nil {
		return
	}

	l = l.With("key", key)
	if err := instance.OnElementRemoved(key); err != nil {
		instance.lastError.Store(err)
		if instance.OnError != nil {
			instance.OnError(key, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("cannot handle element")
		}
	} else {
		l.Debug("element removed")
	}
}

func (instance *Definition) log() log.Logger {
	return log.With("type", instance.typeDescription)
}

func (instance *Definition) logEvent(event string) log.Logger {
	return instance.log().With("event", event)
}

func (instance *Definition) logKey(event string, key string) log.Logger {
	return instance.logEvent(event).With("key", key)
}
