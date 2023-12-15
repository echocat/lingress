package support

import (
	"fmt"
	rt "runtime"
	"time"
)

const (
	timeFormat = "2006-01-02T15:04:05Z"
)

var (
	revision = "latest"
	branch   = "development"
	build    = ""

	runtime = resolveRuntime()
)

// Runtime returns an instance of RuntimeT
func Runtime() RuntimeT {
	return runtime
}

// RuntimeT hold the runtime information about this application. To provide the correct information
// to also enable the usage of:
//
//	Runtime()
//
// ...you have to call the build process with the correct parameters. We have build for that a small utility that
// escapes everything in the right order:
//
//	go run github.com/echocat/lingress/build \
//		-b <branch> \
//		-b <revision> \
//		-o <output> \
//		[<package>]         # Default ./
//
// Example:
//
//	go run github.com/echocat/lingress/build \
//		-g travel-the-pipe \
//		-a some-backend \
//		-b $BRANCH \
//		-b $REVISION \
//		-o out/app \
//		./
type RuntimeT struct {
	Revision  string    `yaml:"revision" json:"revision"`
	Branch    string    `yaml:"branch" json:"branch"`
	Build     time.Time `yaml:"build" json:"build"`
	GoVersion string    `yaml:"goVersion" json:"goVersion"`
	Os        string    `yaml:"os" json:"os"`
	Arch      string    `yaml:"arch" json:"arch"`
}

func (this RuntimeT) LongString() string {
	return fmt.Sprintf(`lingress
 Branch:       %s
 Revision:     %s
 Built:        %s
 Go version:   %s
 OS/Arch:      %s/%s`,
		this.Branch, this.Revision, this.Build, this.GoVersion, this.Os, this.Arch)
}

func (this RuntimeT) String() string {
	return fmt.Sprintf(`%s:%s`,
		this.Branch, this.Revision)
}

func (this RuntimeT) MarshalText() (text []byte, err error) {
	return []byte(this.String()), nil
}

func resolveRuntime() (result RuntimeT) {
	result.Revision = revision
	result.Branch = branch

	//noinspection GoBoolExpressions
	if build == "" {
		result.Build = time.Now()
	} else if t, err := time.Parse(timeFormat, build); err != nil {
		panic(fmt.Sprintf("Illegal build stamp provided (%s): %v", build, err))
	} else {
		result.Build = t
	}

	result.GoVersion = rt.Version()
	result.Os = rt.GOOS
	result.Arch = rt.GOARCH

	return
}
