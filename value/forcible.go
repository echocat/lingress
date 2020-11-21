package value

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
	if v := instance.value; v != nil && v.IsPresent() {
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
