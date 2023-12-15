package value

type String struct {
	value *string
}

func (this String) Get() string {
	return this.GetOr("")
}

func (this String) GetOr(def string) string {
	if v := this.value; v != nil {
		return *v
	}
	return def
}

func (this String) String() string {
	return this.Get()
}

func (this *String) Set(plain string) error {
	if plain == "" {
		*this = String{}
	}
	*this = String{&plain}
	return nil
}

func (this String) IsPresent() bool {
	return this.value != nil
}

type ForcibleString struct {
	Forcible[String, string, *String]
}

func NewForcibleString(init String, forced bool) ForcibleString {
	return ForcibleString{NewForcible[String, string, *String](init, forced)}
}

func (this ForcibleString) Select(target ForcibleString) ForcibleString {
	return ForcibleString{this.Forcible.Select(target.Forcible)}
}
