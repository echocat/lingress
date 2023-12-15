package rules

import (
	"github.com/echocat/lingress/value"
)

type ByHost struct {
	hostFullMatch           map[value.Fqdn]*ByPath
	hostPrefixWildcardMatch map[value.Fqdn]*ByPath
	allHostsMatching        *ByPath

	onAdded        OnAdded
	onRemoved      OnRemoved
	findStrategies []func(host value.Fqdn, path []string) (Rules, error)
}

func NewByHost(onAdded OnAdded, onRemoved OnRemoved) *ByHost {
	result := &ByHost{
		hostFullMatch:           make(map[value.Fqdn]*ByPath),
		hostPrefixWildcardMatch: make(map[value.Fqdn]*ByPath),
		allHostsMatching:        NewByPath(onAdded, onRemoved),

		onAdded:   onAdded,
		onRemoved: onRemoved,
	}
	return result
}

func (this *ByHost) All(consumer func(Rule) error) error {
	for _, v := range this.hostFullMatch {
		if err := v.All(consumer); err != nil {
			return err
		}
	}
	if err := this.allHostsMatching.All(consumer); err != nil {
		return err
	}
	return nil
}

func (this *ByHost) Find(host value.Fqdn, path []string) (Rules, error) {
	var result Rules = rules{}
	for _, strategy := range allFindHostByStrategies {
		candidate, err := strategy(this, host, path)
		if err != nil {
			return nil, err
		}
		if candidate == nil {
			continue
		}

		candidateAny := candidate.AnyFilteredBy(path)
		if candidateAny == nil {
			continue
		}

		resultAny := result.AnyFilteredBy(path)
		if resultAny == nil || len(candidateAny.Path()) > len(resultAny.Path()) {
			result = candidate
		}
	}

	return result, nil
}

var allFindHostByStrategies = []func(instance *ByHost, host value.Fqdn, path []string) (Rules, error){
	findHostByFullMatchStrategy,
	findHostByPrefixWildcardMatchStrategy,
	findHostByAllHostsMatchStrategy,
}

func findHostByFullMatchStrategy(instance *ByHost, host value.Fqdn, path []string) (Rules, error) {
	v, ok := instance.hostFullMatch[host]
	if !ok {
		return nil, nil
	}
	r, err := v.Find(path)
	if err != nil {
		return nil, err
	}
	if r == nil || r.Len() <= 0 {
		return nil, nil
	}
	return r, nil
}

func findHostByPrefixWildcardMatchStrategy(instance *ByHost, host value.Fqdn, path []string) (Rules, error) {
	v, ok := instance.hostPrefixWildcardMatch[host.Parent()]
	if !ok {
		return nil, nil
	}
	r, err := v.Find(path)
	if err != nil {
		return nil, err
	}
	if r == nil || r.Len() <= 0 {
		return nil, nil
	}
	return r, nil
}

func findHostByAllHostsMatchStrategy(instance *ByHost, _ value.Fqdn, path []string) (Rules, error) {
	r, err := instance.allHostsMatching.Find(path)
	if err != nil {
		return nil, err
	}
	if r == nil || r.Len() <= 0 {
		return nil, nil
	}
	return r, nil
}

func (this *ByHost) Put(r Rule) error {
	hostMaybeWithWildcard := r.Host()

	if hostMaybeWithWildcard == "" {
		return this.allHostsMatching.Put(r)
	}

	hadWildcard, host, err := hostMaybeWithWildcard.WithoutWildcard()
	if err != nil {
		return err
	}

	var target map[value.Fqdn]*ByPath
	if hadWildcard {
		target = this.hostPrefixWildcardMatch
	} else {
		target = this.hostFullMatch
	}

	if existing, ok := target[host]; ok {
		return existing.Put(r)
	}

	v := NewByPath(this.onAdded, this.onRemoved)
	target[host] = v
	return v.Put(r)
}

func (this *ByHost) Remove(predicate Predicate) error {
	if err := this.allHostsMatching.Remove(predicate); err != nil {
		return err
	}
	for host, v := range this.hostFullMatch {
		if err := v.Remove(predicate); err != nil {
			return err
		}
		if !v.HasContent() {
			delete(this.hostFullMatch, host)
		}
	}
	for host, v := range this.hostPrefixWildcardMatch {
		if err := v.Remove(predicate); err != nil {
			return err
		}
		if !v.HasContent() {
			delete(this.hostPrefixWildcardMatch, host)
		}
	}
	return nil
}

func (this *ByHost) Clone() *ByHost {
	result := NewByHost(this.onAdded, this.onRemoved)

	result.allHostsMatching = this.allHostsMatching.Clone()
	for k, v := range this.hostFullMatch {
		result.hostFullMatch[k] = v.Clone()
	}
	for k, v := range this.hostPrefixWildcardMatch {
		result.hostPrefixWildcardMatch[k] = v.Clone()
	}

	return result
}
