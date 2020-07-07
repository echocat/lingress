package rules

import (
	"fmt"
	"time"
)

type Value interface {
	String() string
	IsPresent() bool
	Get() interface{}
}

type MutableValue interface {
	Value
	Set(string) error
}

type Bool uint8

const (
	NotDefined = Bool(0)
	False      = Bool(1)
	True       = Bool(2)
)

func NewBool(val bool) Bool {
	if val {
		return True
	}
	return False
}

func (instance Bool) Get() interface{} {
	switch instance {
	case True:
		return true
	case False:
		return false
	}
	return nil
}

func (instance Bool) GetOr(def bool) bool {
	switch instance {
	case True:
		return true
	case False:
		return false
	}
	return def
}

func (instance Bool) String() string {
	switch instance {
	case True:
		return "true"
	case False:
		return "false"
	}
	return ""
}

func (instance *Bool) Set(plain string) error {
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

func (instance Bool) IsPresent() bool {
	return instance == True || instance == False
}

type ForceableBool struct {
	Forceable
}

func NewForceableBool(init Bool, forced bool) ForceableBool {
	val := init
	return ForceableBool{
		Forceable: NewForceable(&val, forced),
	}
}

func (instance ForceableBool) GetOr(def bool) bool {
	switch *(instance.value.(*Bool)) {
	case True:
		return true
	case False:
		return false
	}
	return def
}

func (instance ForceableBool) Select(target ForceableBool) ForceableBool {
	return ForceableBool{
		Forceable: instance.Forceable.Select(target.Forceable),
	}
}

func (instance ForceableBool) Evaluate(other Bool, def bool) bool {
	return instance.Forceable.Evaluate(other, NewBool(def)).(bool)
}

type Forceable struct {
	value  MutableValue
	forced bool
}

func NewForceable(value MutableValue, forced bool) Forceable {
	return Forceable{
		value:  value,
		forced: forced,
	}
}

func (instance Forceable) Get() interface{} {
	return instance.value.Get()
}

func (instance Forceable) Evaluate(other Value, def Value) interface{} {
	if instance.forced {
		if instance.IsPresent() {
			return instance.Get()
		}
		return def.Get()
	}
	if other.IsPresent() {
		return other.Get()
	}
	return def.Get()
}

func (instance Forceable) Select(target Forceable) Forceable {
	if instance.forced {
		return instance
	}
	if target.IsPresent() {
		return target
	}
	return instance
}

func (instance *Forceable) Set(plain string) error {
	forced := false
	if len(plain) > 0 && plain[0] == '!' {
		forced = true
		plain = plain[1:]
	}
	if err := instance.value.Set(plain); err != nil {
		return err
	}
	instance.forced = forced
	return nil
}

func (instance Forceable) String() string {
	result := ""
	if instance.forced {
		result += "!"
	}
	if v := instance.value; v.IsPresent() {
		result += v.String()
	}
	return result
}

func (instance Forceable) IsPresent() bool {
	return instance.value.IsPresent()
}

func (instance Forceable) IsForced() bool {
	return instance.forced
}

type Duration time.Duration

func ParseDuration(plain string) (result Duration, err error) {
	err = result.Set(plain)
	return
}

func (instance Duration) Get() interface{} {
	return instance
}

func (instance Duration) AsSeconds() int {
	return int(time.Duration(instance) / time.Second)
}

func (instance Duration) String() string {
	if instance > 0 {
		return time.Duration(instance).String()
	}
	return ""
}

func (instance *Duration) Set(plain string) error {
	if plain == "" {
		*instance = 0
		return nil
	}

	val, err := time.ParseDuration(plain)
	if err != nil {
		return err
	}

	*instance = Duration(val)
	return nil
}

func (instance Duration) IsPresent() bool {
	return instance > 0
}

type ForceableDuration struct {
	Forceable
}

func NewForceableDuration(init Duration, forced bool) ForceableDuration {
	val := init
	return ForceableDuration{
		Forceable: NewForceable(&val, forced),
	}
}

func (instance ForceableDuration) Evaluate(other Duration, def Duration) Duration {
	return instance.Forceable.Evaluate(other, def).(Duration)
}
