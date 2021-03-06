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

func (instance *ByHost) All(consumer func(Rule) error) error {
	for _, value := range instance.hostFullMatch {
		if err := value.All(consumer); err != nil {
			return err
		}
	}
	if err := instance.allHostsMatching.All(consumer); err != nil {
		return err
	}
	return nil
}

func (instance *ByHost) Find(host value.Fqdn, path []string) (Rules, error) {
	var result Rules = rules{}
	for _, strategy := range allFindHostByStrategies {
		candidate, err := strategy(instance, host, path)
		if err != nil {
			return nil, err
		}
		if candidate == nil {
			continue
		}
		resultAny, candidateAny := result.Any(), candidate.Any()
		if candidateAny == nil {
			continue
		}
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

func (instance *ByHost) Put(r Rule) error {
	hostMaybeWithWildcard := r.Host()

	if hostMaybeWithWildcard == "" {
		return instance.allHostsMatching.Put(r)
	}

	hadWildcard, host, err := hostMaybeWithWildcard.WithoutWildcard()
	if err != nil {
		return err
	}

	var target map[value.Fqdn]*ByPath
	if hadWildcard {
		target = instance.hostPrefixWildcardMatch
	} else {
		target = instance.hostFullMatch
	}

	if existing, ok := target[host]; ok {
		return existing.Put(r)
	}

	value := NewByPath(instance.onAdded, instance.onRemoved)
	target[host] = value
	return value.Put(r)
}

func (instance *ByHost) Remove(predicate Predicate) error {
	if err := instance.allHostsMatching.Remove(predicate); err != nil {
		return err
	}
	for host, v := range instance.hostFullMatch {
		if err := v.Remove(predicate); err != nil {
			return err
		}
		if !v.HasContent() {
			delete(instance.hostFullMatch, host)
		}
	}
	for host, v := range instance.hostPrefixWildcardMatch {
		if err := v.Remove(predicate); err != nil {
			return err
		}
		if !v.HasContent() {
			delete(instance.hostPrefixWildcardMatch, host)
		}
	}
	return nil
}

func (instance *ByHost) Clone() *ByHost {
	result := NewByHost(instance.onAdded, instance.onRemoved)

	result.allHostsMatching = instance.allHostsMatching.Clone()
	for k, v := range instance.hostFullMatch {
		result.hostFullMatch[k] = v.Clone()
	}
	for k, v := range instance.hostPrefixWildcardMatch {
		result.hostPrefixWildcardMatch[k] = v.Clone()
	}

	return result
}
