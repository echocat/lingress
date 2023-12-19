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

func (this rules) Get(i int) Rule {
	return this[i]
}

func (this rules) Len() int {
	if this == nil {
		return 0
	}
	return len(this)
}

func (this rules) Any() Rule {
	if len(this) > 0 {
		return (this)[0]
	}
	return nil
}

func (this rules) AnyFilteredBy(path []string) Rule {
	for _, candidate := range this {
		if candidate.PathType() != PathTypeExact {
			return candidate
		}
		if slices.Equal(path, candidate.Path()) {
			return candidate
		}
	}
	return nil
}

func (this rules) String() string {
	result := ""
	for i, r := range this {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprint(r)
	}
	return result
}
