package support

type Initializable interface {
	Init(stop Channel) error
}
