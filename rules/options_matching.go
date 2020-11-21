package rules

import (
	"github.com/echocat/lingress/value"
	"strings"
)

var _ = RegisterDefaultOptionsPart(&OptionsMatching{})

const (
	optionsMatchingKey = "matching"

	annotationHostsWildcard = "lingress.echocat.org/hosts-wildcard"
)

func OptionsMatchingOf(rule Rule) *OptionsMatching {
	if rule == nil {
		return &OptionsMatching{}
	}
	if v, ok := rule.Options()[optionsMatchingKey].(*OptionsMatching); ok {
		return v
	}
	return &OptionsMatching{}
}

type OptionsMatching struct {
	HostWildcard value.String `json:"wildcardPattern,omitempty"`
}

func (instance OptionsMatching) Name() string {
	return optionsMatchingKey
}

func (instance OptionsMatching) IsRelevant() bool {
	return len(instance.HostWildcard) > 0
}

func (instance *OptionsMatching) Set(annotations Annotations) (err error) {
	if instance.HostWildcard, err = evaluationOptionHostWildcard(annotations); err != nil {
		return
	}
	return
}

func evaluationOptionHostWildcard(annotations map[string]string) (result value.String, err error) {
	return value.String(strings.TrimSpace(annotations[annotationHostsWildcard])), nil
}
