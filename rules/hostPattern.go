package rules

import "strings"

type HostPattern []string

func ParseHostPattern(plain string) (result HostPattern, err error) {
	err = result.Set(plain)
	return
}

func (instance *HostPattern) Set(plain string) error {
	parts := strings.Split(plain, ".")
	*instance = parts
	return nil
}

func (instance HostPattern) String() string {
	return strings.Join(instance, ".")
}

func (instance HostPattern) Matches(test string) bool {
	expected := strings.Split(test, ".")

	if len(expected) != len(instance) {
		return false
	}

	for i, part := range instance {
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

func (instance *HostPatterns) Set(plain string) error {
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

	*instance = result
	return nil
}

func (instance HostPatterns) String() string {
	plains := make([]string, len(instance))
	for i, part := range instance {
		plains[i] = part.String()
	}
	return strings.Join(plains, ",")
}

func (instance HostPatterns) Matches(test string) bool {
	if len(instance) == 0 ||
		(len(instance) == 1 && len(instance[0]) == 1 && instance[0][0] == "*") {
		return true
	}
	for _, candidate := range instance {
		if candidate.Matches(test) {
			return true
		}
	}

	return false
}

func (instance HostPatterns) Get() interface{} {
	return instance
}

func (instance HostPatterns) IsPresent() bool {
	return len(instance) > 0
}

type ForceableHostPatterns struct {
	Forceable
}

func NewForceableHostPatterns(init HostPatterns, forced bool) ForceableHostPatterns {
	val := init
	return ForceableHostPatterns{
		Forceable: NewForceable(&val, forced),
	}
}

func (instance ForceableHostPatterns) Evaluate(other HostPatterns, def HostPatterns) HostPatterns {
	return instance.Forceable.Evaluate(other, def).(HostPatterns)
}
