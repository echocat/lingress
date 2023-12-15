package value

import (
	"time"
)

type Duration struct {
	value *time.Duration
}

func NewDuration(value time.Duration) Duration {
	return Duration{&value}
}

func ParseDuration(plain string) (result Duration, err error) {
	err = result.Set(plain)
	return
}

func (this Duration) Get() time.Duration {
	if v := this.value; v != nil {
		return *v
	}
	return 0
}

func (this Duration) GetOr(def time.Duration) time.Duration {
	if v := this.value; v != nil {
		return *v
	}
	return def
}

func (this Duration) AsSeconds() int {
	return int(this.Get() / time.Second)
}

func (this Duration) String() string {
	if v := this.value; v != nil {
		return v.String()
	}
	return ""
}

func (this *Duration) Set(plain string) error {
	if plain == "" {
		*this = Duration{}
		return nil
	}

	val, err := time.ParseDuration(plain)
	if err != nil {
		return err
	}

	*this = Duration{&val}
	return nil
}

func (this Duration) IsPresent() bool {
	return this.value != nil
}

type ForcibleDuration struct {
	Forcible[Duration, time.Duration, *Duration]
}

func NewForcibleDuration(init Duration, forced bool) ForcibleDuration {
	return ForcibleDuration{NewForcible[Duration, time.Duration, *Duration](init, forced)}
}

func (this ForcibleDuration) Select(target ForcibleDuration) ForcibleDuration {
	return ForcibleDuration{this.Forcible.Select(target.Forcible)}
}
