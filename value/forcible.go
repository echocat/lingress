package value

type Forcible[V Value[T], T any, MV Mutable[V]] struct {
	value  V
	forced bool
}

func NewForcible[V Value[T], T any, MV Mutable[V]](value V, forced bool) Forcible[V, T, MV] {
	return Forcible[V, T, MV]{
		value:  value,
		forced: forced,
	}
}

func (this Forcible[V, T, MV]) Get() T {
	return this.value.Get()
}

func (this Forcible[V, T, MV]) GetOr(def T) T {
	return this.value.GetOr(def)
}

func (this Forcible[V, T, MV]) Evaluate(other V) V {
	var empty V
	return this.EvaluateOr(other, empty)
}

func (this Forcible[V, T, MV]) EvaluateOr(other V, def V) V {
	if this.forced {
		if this.value.IsPresent() {
			return this.value
		}
		return def
	}
	if other.IsPresent() {
		return other
	}
	return def
}

func (this Forcible[V, T, MV]) Select(target Forcible[V, T, MV]) Forcible[V, T, MV] {
	if this.forced {
		return this
	}
	if target.value.IsPresent() {
		return target
	}
	return this
}

func (this *Forcible[V, T, MV]) Set(plain string) error {
	forced := false
	if len(plain) > 0 && plain[0] == '!' {
		forced = true
		plain = plain[1:]
	}
	pv := MV(&this.value)
	if err := pv.Set(plain); err != nil {
		return err
	}
	this.forced = forced
	return nil
}

func (this Forcible[V, T, MV]) String() string {
	result := ""
	if this.forced {
		result += "!"
	}
	if v := this.value; v.IsPresent() {
		result += v.String()
	}
	return result
}

func (this Forcible[V, T, MV]) IsForced() bool {
	return this.forced
}

func (this Forcible[V, T, MV]) IsPresent() bool {
	return this.forced
}
