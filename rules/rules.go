package rules

import (
	"fmt"
	"slices"
)

type Rules interface {
	Get(int) Rule
	Len() int
	Any() Rule
	AnyFilteredBy(path []string) Rule
}

type rules []Rule

func (instance rules) Get(i int) Rule {
	return instance[i]
}

func (instance rules) Len() int {
	if instance == nil {
		return 0
	}
	return len(instance)
}

func (instance rules) Any() Rule {
	if len(instance) > 0 {
		return (instance)[0]
	}
	return nil
}

func (instance rules) AnyFilteredBy(path []string) Rule {
	for _, candidate := range instance {
		if candidate.PathType() != PathTypeExact {
			return candidate
		}
		if slices.Equal(path, candidate.Path()) {
			return candidate
		}
	}
	return nil
}

func (instance rules) String() string {
	result := ""
	for i, r := range instance {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprint(r)
	}
	return result
}
