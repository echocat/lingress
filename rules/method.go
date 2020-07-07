package rules

import (
	"errors"
	"fmt"
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

func (instance *Method) Set(plain string) error {
	candidate := Method(strings.ToUpper(plain))
	if ok := allMethods[candidate]; !ok {
		return fmt.Errorf("%w: %s", ErrIllegalMethod, plain)
	} else {
		*instance = candidate
		return nil
	}
}

func (instance Method) String() string {
	return string(instance)
}

type Methods []Method

func ParseMethods(plain string) (result Methods, err error) {
	err = result.Set(plain)
	return
}

func (instance *Methods) Set(plain string) error {
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
	*instance = result
	return nil
}

func (instance Methods) String() string {
	if len(instance) <= 0 {
		return AllMethods.String()
	}
	plains := make([]string, len(instance))
	for i, part := range instance {
		plains[i] = part.String()
	}
	return strings.Join(plains, ",")
}

func (instance Methods) Matches(test Method) bool {
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

func (instance Methods) Get() interface{} {
	return instance
}

func (instance Methods) IsPresent() bool {
	return len(instance) > 0
}

type ForceableMethods struct {
	Forceable
}

func NewForceableMethods(init Methods, forced bool) ForceableMethods {
	val := init
	return ForceableMethods{
		Forceable: NewForceable(&val, forced),
	}
}

func (instance ForceableMethods) Evaluate(other Methods, def Methods) Methods {
	return instance.Forceable.Evaluate(other, def).(Methods)
}
