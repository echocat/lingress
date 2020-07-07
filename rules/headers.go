package rules

import (
	"strings"
)

type HeaderName string

func ParseHeaderName(plain string) (result HeaderName, err error) {
	err = result.Set(plain)
	return
}

func (instance *HeaderName) Set(plain string) error {
	*instance = HeaderName(strings.ToLower(plain))
	return nil
}

func (instance HeaderName) String() string {
	return string(instance)
}

type HeaderNames []HeaderName

func ParseHeaderNames(plain string) (result HeaderNames, err error) {
	err = result.Set(plain)
	return
}

func (instance *HeaderNames) Set(plain string) error {
	var result HeaderNames
	for _, plainPart := range strings.Split(plain, ",") {
		plainPart = strings.TrimSpace(plainPart)
		if plainPart != "" {
			if part, err := ParseHeaderName(plainPart); err != nil {
				return err
			} else {
				result = append(result, part)
			}
		}
	}
	*instance = result
	return nil
}

func (instance HeaderNames) String() string {
	if len(instance) <= 0 {
		return "*"
	}
	plains := make([]string, len(instance))
	for i, part := range instance {
		plains[i] = part.String()
	}
	return strings.Join(plains, ",")
}

func (instance HeaderNames) Matches(test HeaderName) bool {
	if len(instance) == 0 {
		return true
	}

	for _, candidate := range instance {
		if candidate == test {
			return true
		}
	}

	return false
}

func (instance HeaderNames) Get() interface{} {
	return instance
}

func (instance HeaderNames) IsPresent() bool {
	return len(instance) > 0
}

type ForceableHeaderNames struct {
	Forceable
}

func NewForceableHeaders(init HeaderNames, forced bool) ForceableHeaderNames {
	val := init
	return ForceableHeaderNames{
		Forceable: NewForceable(&val, forced),
	}
}

func (instance ForceableHeaderNames) Evaluate(other HeaderNames, def HeaderNames) HeaderNames {
	return instance.Forceable.Evaluate(other, def).(HeaderNames)
}
