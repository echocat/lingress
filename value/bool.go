package value

import (
	"fmt"
)

type Bool uint8

const (
	UndefinedBool = Bool(0)
	False         = Bool(1)
	True          = Bool(2)
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
		*instance = UndefinedBool
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
