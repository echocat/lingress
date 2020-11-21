package value

type Value interface {
	String() string
	IsPresent() bool
	Get() interface{}
}

type MutableValue interface {
	Value
	Set(string) error
}
