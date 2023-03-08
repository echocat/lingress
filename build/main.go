package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	timeFormat  = "2006-01-02T15:04:05Z"
	basePackage = "github.com/echocat/lingress/support"
)

var (
	groupId = kingpin.Flag("groupId", "").
		Envar("GROUP_ID").
		Short('g').
		Default("").
		String()
	artifactId = kingpin.Flag("artifactId", "").
			Envar("ARTIFACT_ID").
			Short('a').
			Default("").
			String()
	branch = kingpin.Flag("branch", "").
		Envar("BRANCH").
		Short('b').
		Required().
		String()
	revision = kingpin.Flag("revision", "").
			Envar("REVISION").
			Short('r').
			Required().
			String()
	output = kingpin.Flag("output", "").
		Envar("OUTPUT").
		Short('o').
		Required().
		String()
	targetPackage = kingpin.Arg("package", "").
			Envar("PACKAGE").
			Default("./").
			String()
)

func main() {
	support.MustRegisterGlobalFalgs(kingpin.CommandLine, "BUILDER_")
	kingpin.Parse()

	support.Must(os.MkdirAll(filepath.Dir(*output), 0755))

	mustExecute("go", "build", "-ldflags", formatLdFlags(), "-o", *output, *targetPackage)
}

func mustExecute(args ...string) {
	if len(args) <= 0 {
		panic("no arguments provided")
	}
	log.Debugf("Execute: %s", support.QuoteAndJoin(args...))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("command failed [%s]: %v", strings.Join(args, " "), err))
	}
}

func formatLdFlags() string {
	ldFlags := "" +
		fmt.Sprintf(" -X %s.groupId=%s", basePackage, *groupId) +
		fmt.Sprintf(" -X %s.artifactId=%s", basePackage, *artifactId) +
		fmt.Sprintf(" -X %s.branch=%s", basePackage, *branch) +
		fmt.Sprintf(" -X %s.revision=%s", basePackage, *revision) +
		fmt.Sprintf(" -X %s.build=%s", basePackage, time.Now().Format(timeFormat))
	return ldFlags
}
