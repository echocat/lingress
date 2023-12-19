package value

import (
	"github.com/echocat/lingress/value"
	"strings"
)

type HeaderName string

func ParseHeaderName(plain string) (result HeaderName, err error) {
	err = result.Set(plain)
	return
}

func (this *HeaderName) Set(plain string) error {
	*this = HeaderName(strings.ToLower(plain))
	return nil
}

func (this HeaderName) String() string {
	return string(this)
}

type HeaderNames []HeaderName

func ParseHeaderNames(plain string) (result HeaderNames, err error) {
	err = result.Set(plain)
	return
}

func (this *HeaderNames) Set(plain string) error {
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
	*this = result
	return nil
}

func (this HeaderNames) String() string {
	if len(this) <= 0 {
		return "*"
	}
	plains := make([]string, len(this))
	for i, part := range this {
		plains[i] = part.String()
	}
	return strings.Join(plains, ",")
}

func (this HeaderNames) Matches(test HeaderName) bool {
	if len(this) == 0 {
		return true
	}

	for _, candidate := range this {
		if candidate == test {
			return true
		}
	}

	return false
}

func (this HeaderNames) Get() HeaderNames {
	return this
}

func (this HeaderNames) GetOr(def HeaderNames) HeaderNames {
	if len(this) == 0 {
		return def
	}
	return this
}

func (this HeaderNames) IsPresent() bool {
	return len(this) > 0
}

type ForcibleHeaderNames struct {
	value.Forcible[HeaderNames, HeaderNames, *HeaderNames]
}

func NewForcibleHeaders(init HeaderNames, forced bool) ForcibleHeaderNames {
	return ForcibleHeaderNames{value.NewForcible[HeaderNames, HeaderNames, *HeaderNames](init, forced)}
}

func (this ForcibleHeaderNames) Select(target ForcibleHeaderNames) ForcibleHeaderNames {
	return ForcibleHeaderNames{this.Forcible.Select(target.Forcible)}
}
