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

type OnElementChangedFunc func(ref support.ObjectReference, new metav1.Object) error
type OnElementUpdatedFunc func(ref support.ObjectReference, old, new metav1.Object) error
type OnElementRemovedFunc func(ref support.ObjectReference) error
type OnErrorFunc func(ref support.ObjectReference, event string, err error)

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
	this.Logger.Info("Definition store started.")

	go this.informer.Run(support.ToChan(stop))
	stop.Wait()

	this.Logger.Info("Definition store stopped.")
}

func (this *Definition) onClusterElementAdded(new interface{}) {
	l := this.logEvent("elementAdded")

	objSource, ok := new.(support.ObjectReferenceSource)
	if !ok {
		l.With("objectType", reflect.TypeOf(new)).
			Error("Cannot determine reference of an object of type.")
	}
	ref, err := support.NewObjectReferenceOf(objSource)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(new)).
			Error("Cannot determine reference of an object of type.")
	}

	if this.OnElementAdded == nil {
		return
	}

	l = l.With("ref", ref)
	if err := this.OnElementAdded(ref, objSource); err != nil {
		this.lastError.Store(err)
		if this.OnError != nil {
			this.OnError(ref, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("Cannot handle element.")
		}
	} else {
		l.Debug("Element added.")
	}
}

func (this *Definition) onClusterElementUpdated(old interface{}, new interface{}) {
	l := this.logEvent("elementUpdated")

	objSource, ok := new.(support.ObjectReferenceSource)
	if !ok {
		l.With("objectType", reflect.TypeOf(new)).
			Error("Cannot determine reference of an object of type.")
	}
	ref, err := support.NewObjectReferenceOf(objSource)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(new)).
			Error("Cannot determine reference of an object of type.")
	}

	if this.OnElementUpdated == nil {
		return
	}

	l = l.With("ref", ref)
	if err := this.OnElementUpdated(ref, old.(metav1.Object), objSource); err != nil {
		this.lastError.Store(err)
		if this.OnError != nil {
			this.OnError(ref, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("Cannot handle element.")
		}
	} else {
		l.Debug("Element updated.")
	}
}

func (this *Definition) onClusterElementRemoved(old interface{}) {
	l := this.logEvent("elementRemoved")

	objSource, ok := old.(support.ObjectReferenceSource)
	if !ok {
		l.With("objectType", reflect.TypeOf(old)).
			Error("Cannot determine key of an object of type.")
	}
	ref, err := support.NewObjectReferenceOf(objSource)
	if err != nil {
		l.WithError(err).
			With("objectType", reflect.TypeOf(old)).
			Error("Cannot determine reference of an object of type.")
	}

	if this.OnElementRemoved == nil {
		return
	}

	l = l.With("ref", ref)
	if err := this.OnElementRemoved(ref); err != nil {
		this.lastError.Store(err)
		if this.OnError != nil {
			this.OnError(ref, "elementRemoved", err)
		} else {
			l.WithError(err).
				Error("Cannot handle element.")
		}
	} else {
		l.Debug("Element removed.")
	}
}

func (this *Definition) logEvent(event string) log.Logger {
	return this.Logger.With("event", event)
}
