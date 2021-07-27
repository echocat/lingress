package value

type String string

func (instance String) Get() interface{} {
	return instance
}

func (instance String) String() string {
	return string(instance)
}

func (instance *String) Set(plain string) error {
	*instance = String(plain)
	return nil
}

func (instance String) IsPresent() bool {
	return instance != ""
}

type ForcibleString struct {
	Forcible
}

func NewForcibleString(init String, forced bool) ForcibleString {
	val := init
	return ForcibleString{
		Forcible: NewForcible(&val, forced),
	}
}

func (instance ForcibleString) Evaluate(other String, def String) String {
	return instance.Forcible.Evaluate(other, def).(String)
}
