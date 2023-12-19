package rules

import (
	"fmt"
	"github.com/echocat/lingress/value"
	"net"
	"reflect"
	"strings"
	"sync"
)

var (
	optionsPartPrototypes      = map[string]OptionsPart{}
	optionsPartPrototypesMutex = new(sync.RWMutex)

	DefaultOptionsFactory OptionsFactory = newOptions
)

type Options map[string]OptionsPart

type OptionsFactory func() Options

func newOptions() Options {
	optionsPartPrototypesMutex.RLock()
	defer optionsPartPrototypesMutex.RUnlock()

	result := make(Options, len(optionsPartPrototypes))
	for _, prototype := range optionsPartPrototypes {
		result[prototype.Name()] = newOptionsPartBy(prototype)
	}

	return result
}

type OptionsPart interface {
	Name() string
	IsRelevant() bool
	Set(annotations Annotations) error
}

type Annotations map[string]string

func (this Options) IsRelevant() bool {
	if this == nil {
		return false
	}

	for _, part := range this {
		if ok := part.IsRelevant(); ok {
			return true
		}
	}

	return false
}

func (this *Options) Set(annotations Annotations) error {
	if this == nil {
		return nil
	}

	for _, part := range *this {
		if err := part.Set(annotations); err != nil {
			return err
		}
	}

	return nil
}

func AnnotationIsBool(name, v string) (value.Bool, error) {
	switch v {
	case "true":
		return value.True(), nil
	case "false":
		return value.False(), nil
	case "":
		return value.UndefinedBool(), nil
	default:
		return value.Bool{}, fmt.Errorf("illegal boolean value for annotation %s: %s", name, v)
	}
}

func AnnotationIsForcibleBool(name, v string) (result value.ForcibleBool, err error) {
	result = value.NewForcibleBool(value.False(), false)
	if err := result.Set(v); err != nil {
		return value.ForcibleBool{}, fmt.Errorf("illegal boolean value for annotation %s: %s", name, v)
	}
	return
}

func AnnotationAddresses(name, value string) (result []Address, err error) {
	for _, candidate := range strings.Split(value, ",") {
		if strings.Contains(candidate, "/") {
			if _, n, pErr := net.ParseCIDR(candidate); pErr != nil {
				return nil, fmt.Errorf("'%s' is an illegal CIDR for annotation '%s': %v", candidate, name, pErr)
			} else {
				result = append(result, &networkAddress{n})
			}
		} else {
			if ips, pErr := net.LookupIP(candidate); pErr != nil {
				return nil, fmt.Errorf("'%s' is an illegal address for annotation '%s': %v", candidate, name, pErr)
			} else {
				for _, ip := range ips {
					result = append(result, ipAddress(ip))
				}
			}
		}
	}
	return
}

func newOptionsPartBy(prototype OptionsPart) (out OptionsPart) {
	t := reflect.TypeOf(prototype)
	var v reflect.Value
	if t.Kind() == reflect.Ptr {
		v = reflect.New(t.Elem())
	} else {
		v = reflect.New(t).Elem()
	}

	return v.Interface().(OptionsPart)
}

func RegisterDefaultOptionsPart(prototype OptionsPart) OptionsPart {
	optionsPartPrototypesMutex.Lock()
	defer optionsPartPrototypesMutex.Unlock()

	name := prototype.Name()
	if _, exists := optionsPartPrototypes[name]; exists {
		panic(fmt.Sprintf("multiple registrations of object part '%v'", name))
	}

	optionsPartPrototypes[name] = prototype

	return prototype
}
