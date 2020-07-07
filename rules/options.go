package rules

import (
	"fmt"
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

func (instance Options) IsRelevant() bool {
	if instance == nil {
		return false
	}

	for _, part := range instance {
		if ok := part.IsRelevant(); ok {
			return true
		}
	}

	return false
}

func (instance *Options) Set(annotations Annotations) error {
	if instance == nil {
		return nil
	}

	for _, part := range *instance {
		if err := part.Set(annotations); err != nil {
			return err
		}
	}

	return nil
}

func AnnotationIsTrue(name, value string) (Bool, error) {
	if value == "true" {
		return True, nil
	}
	if value == "false" {
		return False, nil
	}
	return 0, fmt.Errorf("illegal boolean value for annotation %s: %s", name, value)
}

func AnnotationIsForceableBool(name, value string) (result ForceableBool, err error) {
	result = NewForceableBool(False, false)
	if err := result.Set(value); err != nil {
		return ForceableBool{}, fmt.Errorf("illegal boolean value for annotation %s: %s", name, value)
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
