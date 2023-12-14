package support

import (
	"fmt"
	"os"
	"path/filepath"
	rt "runtime"
	"time"
)

const (
	timeFormat = "2006-01-02T15:04:05Z"
)

var (
	groupId    = ""
	artifactId = ""
	revision   = "latest"
	branch     = "development"
	build      = ""

	runtime    = resolveRuntime()
	executable = resolveExecutable()
)

// Runtime returns an instance of RuntimeT
func Runtime() RuntimeT {
	return runtime
}

// Executable returns the actual name of the executable
func Executable() string {
	return executable
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
//		[-g <groupId>] \    # Default: ""
//		-a <artifactId> \
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
	GroupId    string    `yaml:"groupId" json:"groupId"`
	ArtifactId string    `yaml:"artifactId" json:"artifactId"`
	Revision   string    `yaml:"revision" json:"revision"`
	Branch     string    `yaml:"branch" json:"branch"`
	Build      time.Time `yaml:"build" json:"build"`
	GoVersion  string    `yaml:"goVersion" json:"goVersion"`
	Os         string    `yaml:"os" json:"os"`
	Arch       string    `yaml:"arch" json:"arch"`
}

func (instance RuntimeT) Name() string {
	g := instance.GroupId
	a := instance.ArtifactId
	if g == "" {
		return a
	}
	return g + "/" + a
}

func (instance RuntimeT) LongString() string {
	return fmt.Sprintf(`%s
 Branch:       %s
 Revision:     %s
 Built:        %s
 Go version:   %s
 OS/Arch:      %s/%s`,
		instance.Name(), instance.Branch, instance.Revision, instance.Build, instance.GoVersion, instance.Os, instance.Arch)
}

func (instance RuntimeT) String() string {
	return fmt.Sprintf(`%s:%s`,
		instance.Branch, instance.Revision)
}

func (instance RuntimeT) MarshalText() (text []byte, err error) {
	return []byte(instance.String()), nil
}

func resolveRuntime() (result RuntimeT) {
	result.GroupId = groupId
	result.ArtifactId = artifactId
	if result.ArtifactId == "" {
		if fallback, err := os.Executable(); err != nil {
			result.ArtifactId = filepath.Base(os.Args[0])
		} else {
			result.ArtifactId = filepath.Base(fallback)
		}
	}

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

func resolveExecutable() string {
	if result, err := os.Executable(); err != nil {
		return os.Args[0]
	} else {
		return result
	}
}
