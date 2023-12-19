package value

import (
	"errors"
	"fmt"
	"github.com/echocat/lingress/value"
	"net/http"
	"strings"
)

type Method string

const (
	MethodGet     = Method(http.MethodGet)
	MethodHead    = Method(http.MethodHead)
	MethodPost    = Method(http.MethodPost)
	MethodPut     = Method(http.MethodPut)
	MethodPatch   = Method(http.MethodPatch)
	MethodDelete  = Method(http.MethodDelete)
	MethodConnect = Method(http.MethodConnect)
	MethodOptions = Method(http.MethodOptions)
	MethodTrace   = Method(http.MethodTrace)
)

var (
	ErrIllegalMethod = errors.New("illegal method")

	AllMethods = Methods{
		MethodGet,
		MethodHead,
		MethodPost,
		MethodPut,
		MethodPatch,
		MethodDelete,
		MethodConnect,
		MethodOptions,
		MethodTrace,
	}

	allMethods = func(in []Method) map[Method]bool {
		result := make(map[Method]bool, len(in))
		for _, method := range in {
			result[method] = true
		}
		return result
	}(AllMethods)
)

func ParseMethod(plain string) (result Method, err error) {
	err = result.Set(plain)
	return
}

func (this *Method) Set(plain string) error {
	candidate := Method(strings.ToUpper(plain))
	if ok := allMethods[candidate]; !ok {
		return fmt.Errorf("%w: %s", ErrIllegalMethod, plain)
	} else {
		*this = candidate
		return nil
	}
}

func (this Method) String() string {
	return string(this)
}

type Methods []Method

func ParseMethods(plain string) (result Methods, err error) {
	err = result.Set(plain)
	return
}

func (this *Methods) Set(plain string) error {
	var result Methods
	for _, plainPart := range strings.Split(plain, ",") {
		plainPart = strings.TrimSpace(plainPart)
		if plainPart != "" {
			if part, err := ParseMethod(plainPart); err != nil {
				return err
			} else {
				result = append(result, part)
			}
		}
	}
	*this = result
	return nil
}

func (this Methods) String() string {
	if len(this) <= 0 {
		return AllMethods.String()
	}
	plains := make([]string, len(this))
	for i, part := range this {
		plains[i] = part.String()
	}
	return strings.Join(plains, ",")
}

func (this Methods) Matches(test Method) bool {
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

func (this Methods) Get() Methods {
	return this
}

func (this Methods) GetOr(def Methods) Methods {
	if len(this) == 0 {
		return def
	}
	return this
}

func (this Methods) IsPresent() bool {
	return len(this) > 0
}

type ForcibleMethods struct {
	value.Forcible[Methods, Methods, *Methods]
}

func NewForcibleMethods(init Methods, forced bool) ForcibleMethods {
	return ForcibleMethods{value.NewForcible[Methods, Methods, *Methods](init, forced)}
}

func (this ForcibleMethods) Select(target ForcibleMethods) ForcibleMethods {
	return ForcibleMethods{this.Forcible.Select(target.Forcible)}
}
