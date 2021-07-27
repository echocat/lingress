package rules

import (
	"encoding/json"
	"fmt"
	"github.com/echocat/lingress/value"
	"net"
	"strings"
)

type Rule interface {
	Host() value.WildcardSupportingFqdn
	Path() []string
	Source() SourceReference
	Backend() net.Addr
	Options() Options
	Statistics() *Statistics
}

type Predicate func(path []string, r Rule) bool

func PredicateBySource(reference SourceReference) Predicate {
	return func(path []string, r Rule) bool {
		return reference.Equals(r.Source())
	}
}

type OnAdded func(path []string, r Rule)
type OnRemoved func(path []string, r Rule)

type rule struct {
	host       value.WildcardSupportingFqdn
	path       []string
	source     SourceReference
	backend    net.Addr
	options    Options
	statistics *Statistics
}

func NewRule(host value.WildcardSupportingFqdn, path []string, source SourceReference, backend net.Addr, options Options) Rule {
	return &rule{
		host:       host,
		path:       path,
		source:     source,
		backend:    backend,
		options:    options,
		statistics: &Statistics{},
	}
}

func (instance *rule) clone() *rule {
	r := *instance
	return &r
}

func (instance *rule) Host() value.WildcardSupportingFqdn {
	return instance.host
}

func (instance *rule) Path() []string {
	return instance.path
}

func (instance *rule) Source() SourceReference {
	return instance.source
}

func (instance *rule) Backend() net.Addr {
	return instance.backend
}

func (instance *rule) Options() Options {
	return instance.options
}

func (instance *rule) Statistics() *Statistics {
	return instance.statistics
}

func (instance *rule) String() string {
	return fmt.Sprintf("(%v) %s/%s -> %v", instance.Source(), instance.Host(), strings.Join(instance.Path(), "/"), instance.Backend())
}

func (instance *rule) MarshalJSON() ([]byte, error) {
	buf := make(map[string]string)
	buf["host"] = instance.host.String()
	buf["path"] = "/" + strings.Join(instance.Path(), "/")
	buf["source"] = instance.Source().String()
	buf["backend"] = instance.Backend().String()
	return json.Marshal(buf)
}
