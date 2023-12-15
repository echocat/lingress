package value

type Value[T any] interface {
	String() string
	IsPresent() bool
	GetOr(def T) T
	Get() T
}

type Mutable[T any] interface {
	Set(string) error
	*T
}
