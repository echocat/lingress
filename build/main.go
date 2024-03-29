package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	_ "github.com/echocat/lingress/support/slf4g_native"
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
	"regexp"
)

var (
	sanitizeBranchNameRegex = regexp.MustCompile(`[^a-zA-Z0-9-.]+`)
)

func newCmd() cmd {
	return cmd{
		outputPrefix:  "dist/lingress",
		dockerCommand: "docker",
		withImages:    true,
	}
}

type cmd struct {
	branch       string
	revision     string
	outputPrefix string

	isLatest      bool
	dockerCommand string

	withTests  bool
	withBuild  bool
	withImages bool
	withDeploy bool
}

func (this *cmd) registerFlags(app *kingpin.Application) {
	app.Flag("branch", "").
		Envar("GITHUB_REF_NAME").
		Short('c').
		Required().
		StringVar(&this.branch)
	app.Flag("revision", "").
		Envar("GITHUB_SHA").
		Short('r').
		Required().
		StringVar(&this.revision)
	app.Flag("outputPrefix", "").
		Envar("BUILDER_OUTPUT_PREFIX").
		Short('o').
		PlaceHolder(this.outputPrefix).
		StringVar(&this.outputPrefix)
	app.Flag("isLatest", "").
		Envar("BUILDER_IS_LATEST").
		BoolVar(&this.isLatest)
	app.Flag("dockerCommand", "").
		Envar("BUILDER_DOCKER_COMMAND").
		PlaceHolder(this.dockerCommand).
		StringVar(&this.dockerCommand)
	app.Flag("test", "").
		PlaceHolder(fmt.Sprint(this.withTests)).
		Envar("BUILDER_WITH_TESTS").
		Short('t').
		BoolVar(&this.withTests)
	app.Flag("build", "").
		PlaceHolder(fmt.Sprint(this.withBuild)).
		Envar("BUILDER_WITH_BUILD").
		Short('b').
		BoolVar(&this.withBuild)
	app.Flag("withImages", "").
		PlaceHolder(fmt.Sprint(this.withImages)).
		Envar("BUILDER_WITH_IMAGES").
		Short('i').
		BoolVar(&this.withImages)
	app.Flag("deploy", "").
		Envar("BUILDER_WITH_DEPLOY").
		PlaceHolder(fmt.Sprint(this.withDeploy)).
		Short('d').
		BoolVar(&this.withDeploy)
}

func (this *cmd) mustExecute() {
	this.branch = sanitizeBranchNameRegex.ReplaceAllString(this.branch, "-")
	if this.withTests {
		this.mustTest()
	}
	if this.withBuild {
		this.mustBuild()
	}
	if this.withDeploy {
		this.mustDeploy()
	}
}

func main() {
	vp := value.NewProvider(native.DefaultProvider)
	kingpin.Flag("log.level", "").
		Envar("BUILDER_LOG_LEVEL").
		SetValue(&vp.Level)
	kingpin.Flag("log.format", "").
		Default("text").
		Envar("BUILDER_LOG_FORMAT").
		SetValue(&vp.Consumer.Formatter)
	kingpin.Flag("log.color", "").
		Default("auto").
		Envar("BUILDER_LOG_COLOR").
		SetValue(vp.Consumer.Formatter.ColorMode)

	cmd := newCmd()
	cmd.registerFlags(kingpin.CommandLine)

	kingpin.Parse()

	cmd.mustExecute()
}
