package support

import (
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pr "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
	"strings"
)

type ObjectReference interface {
	ApiVersion() string
	Kind() string
	Namespace() string
	Name() string
	String() string
	ShortString() string
	Equals(ObjectReference) bool

	WithNamespace(string) ObjectReference
	WithName(string) ObjectReference
	WithGroupVersionKind(schema.GroupVersionKind) ObjectReference
	WithApiVersionAndKind(apiVersion, king string) ObjectReference
	GroupVersionKind() schema.GroupVersionKind
}

type ObjectReferenceSource interface {
	v1.Object
	pr.Object
}

func NewObjectReferenceOf(src ObjectReferenceSource) (ObjectReference, error) {
	gvk := src.GetObjectKind().GroupVersionKind()
	if gvk.Group == "" && gvk.Version == "" {
		candidates, _, err := scheme.Scheme.ObjectKinds(src)
		if err != nil {
			return nil, fmt.Errorf("cannot determine ObjectVersionKind for %v: %w", reflect.TypeOf(src), err)
		}
		if len(candidates) <= 0 {
			return nil, fmt.Errorf("cannot determine ObjectVersionKind for %v", reflect.TypeOf(src))
		}
		gvk = candidates[0]
	}
	apiVersion, kind := gvkToApiVersionAndKind(&gvk)

	return objectReference{
		apiVersion: apiVersion,
		kind:       kind,
		namespace:  src.GetNamespace(),
		name:       src.GetName(),
	}, nil
}

func gvkToApiVersionAndKind(gvk *schema.GroupVersionKind) (apiVersion, kind string) {
	apiVersion = strings.Clone(gvk.Group)
	if v := gvk.Version; v != "" {
		if apiVersion != "" {
			apiVersion += "/"
		}
		apiVersion += v
	}
	kind = strings.Clone(gvk.Kind)
	return
}

type objectReference struct {
	apiVersion string
	kind       string
	namespace  string
	name       string
}

func (this objectReference) Namespace() string {
	return this.namespace
}

func (this objectReference) Name() string {
	return this.name
}

func (this objectReference) ApiVersion() string {
	return this.apiVersion
}

func (this objectReference) Kind() string {
	return this.kind
}

func (this objectReference) WithGroupVersionKind(gvk schema.GroupVersionKind) ObjectReference {
	apiVersion, kind := gvkToApiVersionAndKind(&gvk)
	return objectReference{
		apiVersion: apiVersion,
		kind:       kind,
		namespace:  strings.Clone(this.namespace),
		name:       strings.Clone(this.name),
	}
}

func (this objectReference) WithApiVersionAndKind(apiVersion, kind string) ObjectReference {
	return objectReference{
		apiVersion: strings.Clone(apiVersion),
		kind:       strings.Clone(kind),
		namespace:  strings.Clone(this.namespace),
		name:       strings.Clone(this.name),
	}
}

func (this objectReference) WithNamespace(v string) ObjectReference {
	return objectReference{
		apiVersion: strings.Clone(this.apiVersion),
		kind:       strings.Clone(this.kind),
		namespace:  strings.Clone(v),
		name:       strings.Clone(this.name),
	}
}

func (this objectReference) WithName(v string) ObjectReference {
	return objectReference{
		apiVersion: strings.Clone(this.apiVersion),
		kind:       strings.Clone(this.kind),
		namespace:  strings.Clone(this.namespace),
		name:       strings.Clone(v),
	}
}

func (this objectReference) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(this.apiVersion, this.kind)
}

func (this objectReference) String() string {
	return this.Kind() + ":" + this.ShortString()
}

func (this objectReference) ShortString() string {
	if v := this.namespace; v != "" {
		return v + "/" + this.name
	}
	return strings.Clone(this.name)
}

func (this objectReference) AsShortStringer() fmt.Stringer {
	return objectReferenceShortStringer(this)
}

type objectReferenceShortStringer objectReference

func (this objectReferenceShortStringer) String() string {
	return objectReference(this).ShortString()
}

func (this objectReference) Equals(o ObjectReference) bool {
	if o == nil {
		return false
	}

	return o.ApiVersion() == this.ApiVersion() &&
		o.Kind() == this.Kind() &&
		o.Namespace() == this.Namespace() &&
		o.Name() == this.Name()
}
