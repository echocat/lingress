package value

import (
	"fmt"
	"github.com/echocat/lingress/support"
)

type Bool struct {
	value *bool
}

func UndefinedBool() Bool {
	return Bool{}
}

func False() Bool {
	return Bool{support.AsPtr(false)}
}

func True() Bool {
	return Bool{support.AsPtr(true)}
}

func (this Bool) Get() bool {
	if this.value == nil {
		return false
	}
	return *this.value
}

func (this Bool) GetOr(def bool) bool {
	if this.value == nil {
		return def
	}
	return *this.value
}

func (this Bool) String() string {
	if this.value == nil {
		return ""
	}
	if *this.value {
		return "true"
	}
	return "false"
}

func (this *Bool) Set(plain string) error {
	switch plain {
	case "true":
		*this = True()
		return nil
	case "false":
		*this = False()
		return nil
	case "":
		*this = UndefinedBool()
		return nil
	}
	return fmt.Errorf("illegal value: %s", plain)
}

func (this Bool) IsPresent() bool {
	return this.value != nil
}

type ForcibleBool struct {
	Forcible[Bool, bool, *Bool]
}

func NewForcibleBool(init Bool, forced bool) ForcibleBool {
	return ForcibleBool{NewForcible[Bool, bool, *Bool](init, forced)}
}

func (this ForcibleBool) Select(target ForcibleBool) ForcibleBool {
	return ForcibleBool{this.Forcible.Select(target.Forcible)}
}
