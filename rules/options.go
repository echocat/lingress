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

func AnnotationIsTrue(name, value string) (OptionalBool, error) {
	if value == "true" {
		return True, nil
	}
	if value == "false" {
		return False, nil
	}
	return 0, fmt.Errorf("illegal boolean value for annotation %s: %s", name, value)
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

type OptionalBool uint8

const (
	NotDefined = OptionalBool(0)
	False      = OptionalBool(1)
	True       = OptionalBool(2)
)

func (instance OptionalBool) IsEnabled(def bool) bool {
	switch instance {
	case True:
		return true
	case False:
		return false
	}
	return def
}

func (instance OptionalBool) IsEnabledOrForced(def ForceableBool) bool {
	if def.Forced {
		return def.Value
	}
	switch instance {
	case True:
		return true
	case False:
		return false
	}
	return def.Value
}

func (instance OptionalBool) String() string {
	switch instance {
	case True:
		return "true"
	case False:
		return "false"
	}
	return ""
}

func (instance *OptionalBool) Set(plain string) error {
	switch plain {
	case "true":
		*instance = True
		return nil
	case "false":
		*instance = False
		return nil
	case "":
		*instance = NotDefined
		return nil
	}
	return fmt.Errorf("illegal value: %s", plain)
}

type ForceableBool struct {
	Value  bool
	Forced bool
}

func (instance ForceableBool) String() string {
	if instance.Forced {
		if instance.Value {
			return "!true"
		} else {
			return "!false"
		}
	} else {
		if instance.Value {
			return "true"
		} else {
			return "false"
		}
	}
}

func (instance *ForceableBool) Set(plain string) error {
	switch plain {
	case "true":
		*instance = ForceableBool{
			Value:  true,
			Forced: false,
		}
		return nil
	case "false":
		*instance = ForceableBool{
			Value:  true,
			Forced: false,
		}
		return nil
	case "!true":
		*instance = ForceableBool{
			Value:  true,
			Forced: true,
		}
		return nil
	case "!false":
		*instance = ForceableBool{
			Value:  true,
			Forced: true,
		}
		return nil
	default:
		return fmt.Errorf("illegal value: %s", plain)
	}
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
