package value

import (
	"time"
)

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
