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
	Logger   log.Logger

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

func newDefinition(typeDescription string, informer cache.SharedInformer, logger log.Logger) (*Definition, error) {
	return &Definition{
		typeDescription: typeDescription,
		informer:        informer,
		Logger:          logger.With("type", typeDescription),
	}, nil
}

func (this *Definition) SetInformer(informer cache.SharedInformer) {
	this.informer = informer
}

func (this *Definition) Init(stop support.Channel) error {
	if this.informer == nil {
		panic(fmt.Sprintf("definition %s has no informer", this.typeDescription))
	}

	_, err := this.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    this.onClusterElementAdded,
		UpdateFunc: this.onClusterElementUpdated,
		DeleteFunc: this.onClusterElementRemoved,
	})
	if err != nil {
		return fmt.Errorf("creation of event handler for %s failed: %w", this.typeDescription, err)
	}

	go this.Run(stop)

	if !cache.WaitForCacheSync(support.ToChan(stop), this.HasSynced) {
		stop.Broadcast()
		return fmt.Errorf("initial %s synchronization failed", this.typeDescription)
	}
	if err := this.lastError.Load(); err != nil {
		stop.Broadcast()
		return fmt.Errorf("initial %s synchronization failed: %w", this.typeDescription, err.(error))
	}
	return nil
}

func (this *Definition) HasSynced() bool {
	return this.informer.HasSynced()
}

func (this *Definition) Run(stop support.Channel) {
	defer runtime.HandleCrash()
	this.Logger.Info("definition store started")

	go this.informer.Run(support.ToChan(stop))
	stop.Wait()

	this.Logger.Info("definition store stopped")
}

func (this *Definition) onClusterElementAdded(new interface{}) {
	l := this.logEvent("elementAdded")

	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(new).String()).
			Error("cannot determine key of an object of type")
	}

	if this.OnElementAdded == nil {
		return
	}

	l = l.With("key", key)
	if err := this.OnElementAdded(key, new.(metav1.Object)); err != nil {
		this.lastError.Store(err)
		if this.OnError != nil {
			this.OnError(key, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("cannot handle element")
		}
	} else {
		l.Debug("element added")
	}
}

func (this *Definition) onClusterElementUpdated(old interface{}, new interface{}) {
	l := this.logEvent("elementUpdated")

	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(new).String()).
			Error("cannot determine key of an object of type")
	}

	if this.OnElementUpdated == nil {
		return
	}

	l = l.With("key", key)
	if err := this.OnElementUpdated(key, old.(metav1.Object), new.(metav1.Object)); err != nil {
		this.lastError.Store(err)
		if this.OnError != nil {
			this.OnError(key, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("cannot handle element")
		}
	} else {
		l.Debug("element updated")
	}
}

func (this *Definition) onClusterElementRemoved(old interface{}) {
	l := this.logEvent("elementRemoved")

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(old)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(old).String()).
			Error("cannot determine key of an object of type")
	}

	if this.OnElementRemoved == nil {
		return
	}

	l = l.With("key", key)
	if err := this.OnElementRemoved(key); err != nil {
		this.lastError.Store(err)
		if this.OnError != nil {
			this.OnError(key, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("cannot handle element")
		}
	} else {
		l.Debug("element removed")
	}
}

func (this *Definition) logEvent(event string) log.Logger {
	return this.Logger.With("event", event)
}

func (this *Definition) logKey(event string, key string) log.Logger {
	return this.logEvent(event).With("key", key)
}
