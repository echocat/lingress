package value

import (
	"github.com/echocat/lingress/value"
	"strings"
)

type HostPattern []string

func ParseHostPattern(plain string) (result HostPattern, err error) {
	err = result.Set(plain)
	return
}

func (this *HostPattern) Set(plain string) error {
	parts := strings.Split(plain, ".")
	*this = parts
	return nil
}

func (this HostPattern) String() string {
	return strings.Join(this, ".")
}

func (this HostPattern) Matches(test string) bool {
	expected := strings.Split(test, ".")

	if len(expected) != len(this) {
		return false
	}

	for i, part := range this {
		if part != "*" && part != expected[i] {
			return false
		}
	}

	return true
}

type HostPatterns []HostPattern

func ParseHostPatterns(plain string) (result HostPatterns, err error) {
	err = result.Set(plain)
	return
}

func (this *HostPatterns) Set(plain string) error {
	plain = strings.TrimSpace(plain)

	var result HostPatterns

	if plain != "*" {
		for _, plainPart := range strings.Split(plain, ",") {
			plainPart = strings.TrimSpace(plainPart)
			if plainPart != "" {
				if part, err := ParseHostPattern(plainPart); err != nil {
					return err
				} else {
					result = append(result, part)
				}
			}
		}
	}

	*this = result
	return nil
}

func (this HostPatterns) String() string {
	plains := make([]string, len(this))
	for i, part := range this {
		plains[i] = part.String()
	}
	return strings.Join(plains, ",")
}

func (this HostPatterns) Matches(test string) bool {
	if len(this) == 0 ||
		(len(this) == 1 && len(this[0]) == 1 && this[0][0] == "*") {
		return true
	}
	for _, candidate := range this {
		if candidate.Matches(test) {
			return true
		}
	}

	return false
}

func (this HostPatterns) Get() HostPatterns {
	return this
}

func (this HostPatterns) GetOr(def HostPatterns) HostPatterns {
	if len(this) == 0 {
		return def
	}
	return this
}

func (this HostPatterns) IsPresent() bool {
	return len(this) > 0
}

type ForcibleHostPatterns struct {
	value.Forcible[HostPatterns, HostPatterns, *HostPatterns]
}

func NewForcibleHostPatterns(init HostPatterns, forced bool) ForcibleHostPatterns {
	return ForcibleHostPatterns{value.NewForcible[HostPatterns, HostPatterns, *HostPatterns](init, forced)}
}

func (this ForcibleHostPatterns) Select(target ForcibleHostPatterns) ForcibleHostPatterns {
	return ForcibleHostPatterns{this.Forcible.Select(target.Forcible)}
}
