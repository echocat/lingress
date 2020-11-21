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

type ForcibleBool struct {
	Forcible
}

func NewForcibleBool(init Bool, forced bool) ForcibleBool {
	val := init
	return ForcibleBool{
		Forcible: NewForcible(&val, forced),
	}
}

func (instance ForcibleBool) GetOr(def bool) bool {
	switch *(instance.value.(*Bool)) {
	case True:
		return true
	case False:
		return false
	}
	return def
}

func (instance ForcibleBool) Select(target ForcibleBool) ForcibleBool {
	return ForcibleBool{
		Forcible: instance.Forcible.Select(target.Forcible),
	}
}

func (instance ForcibleBool) Evaluate(other Bool, def bool) bool {
	return instance.Forcible.Evaluate(other, NewBool(def)).(bool)
}

type Forcible struct {
	value  MutableValue
	forced bool
}

func NewForcible(value MutableValue, forced bool) Forcible {
	return Forcible{
		value:  value,
		forced: forced,
	}
}

func (instance Forcible) Get() interface{} {
	return instance.value.Get()
}

func (instance Forcible) Evaluate(other Value, def Value) interface{} {
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

func (instance Forcible) Select(target Forcible) Forcible {
	if instance.forced {
		return instance
	}
	if target.IsPresent() {
		return target
	}
	return instance
}

func (instance *Forcible) Set(plain string) error {
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

func (instance Forcible) String() string {
	result := ""
	if instance.forced {
		result += "!"
	}
	if v := instance.value; v.IsPresent() {
		result += v.String()
	}
	return result
}

func (instance Forcible) IsPresent() bool {
	return instance.value != nil && instance.value.IsPresent()
}

func (instance Forcible) IsForced() bool {
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

type ForcibleDuration struct {
	Forcible
}

func NewForcibleDuration(init Duration, forced bool) ForcibleDuration {
	val := init
	return ForcibleDuration{
		Forcible: NewForcible(&val, forced),
	}
}

func (instance ForcibleDuration) Evaluate(other Duration, def Duration) Duration {
	return instance.Forcible.Evaluate(other, def).(Duration)
}
