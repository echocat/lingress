package rules

import (
	"encoding/json"
	"fmt"
	"github.com/echocat/lingress/rules/tree"
	"github.com/echocat/lingress/value"
	"net"
	"strings"
)

type Rule interface {
	Host() value.WildcardSupportingFqdn
	Path() []string
	PathType() PathType
	Source() SourceReference
	Backend() net.Addr
	Options() Options
	Statistics() *Statistics

	tree.Cloneable[Rule]
}

type PathType uint8

const (
	PathTypeExact PathType = iota
	PathTypePrefix
)

func (this PathType) String() string {
	switch this {
	case PathTypeExact:
		return "Exact"
	case PathTypePrefix:
		return "Prefix"
	default:
		return fmt.Sprintf("Unknown-%d", this)
	}
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
	pathType   PathType
	source     SourceReference
	backend    net.Addr
	options    Options
	statistics *Statistics
}

func NewRule(host value.WildcardSupportingFqdn, path []string, pathType PathType, source SourceReference, backend net.Addr, options Options) Rule {
	return &rule{
		host:       host,
		path:       path,
		pathType:   pathType,
		source:     source,
		backend:    backend,
		options:    options,
		statistics: &Statistics{},
	}
}

func (this *rule) clone() *rule {
	r := *this
	return &r
}

func (this *rule) Clone() Rule {
	return this.clone()
}

func (this *rule) Host() value.WildcardSupportingFqdn {
	return this.host
}

func (this *rule) Path() []string {
	return this.path
}

func (this *rule) PathType() PathType {
	return this.pathType
}

func (this *rule) Source() SourceReference {
	return this.source
}

func (this *rule) Backend() net.Addr {
	return this.backend
}

func (this *rule) Options() Options {
	return this.options
}

func (this *rule) Statistics() *Statistics {
	return this.statistics
}

func (this *rule) String() string {
	return fmt.Sprintf("(%v) %s/%s -> %v", this.Source(), this.Host(), strings.Join(this.Path(), "/"), this.Backend())
}

func (this *rule) MarshalJSON() ([]byte, error) {
	buf := make(map[string]string)
	buf["host"] = this.host.String()
	buf["path"] = "/" + strings.Join(this.Path(), "/")
	buf["source"] = this.Source().String()
	buf["backend"] = this.Backend().String()
	return json.Marshal(buf)
}
