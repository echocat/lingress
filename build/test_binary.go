package main

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (this *cmd) mustTestGoCode(t target) {
	support.Must(os.MkdirAll("var", 0755))
	mustExecuteTo(func(cmd *exec.Cmd) {
		cmd.Env = append(os.Environ(), "GOOS="+t.os, "GOARCH="+t.arch, "CGO_ENABLED=0")
	}, os.Stderr, os.Stdout, "go", "test", "-v", "-covermode", "atomic", "-coverprofile=var/profile.cov", "./...")
}

func (this *cmd) mustTestBinary(t target) {
	started := time.Now()
	l := log.
		With("os", t.os).
		With("arch", t.arch)
	l.Debug("Test binary...")

	this.mustTestBinaryByExpectingResponse(t, `Branch:       TEST`+this.branch+`TEST`, t.outputName(this.outputPrefix), "--version")
	this.mustTestBinaryByExpectingResponse(t, `Revision:     TEST`+this.revision+`TEST`, t.outputName(this.outputPrefix), "--version")

	l.With("duration", time.Now().Sub(started).Truncate(time.Millisecond)).
		Info("Binary tested.")
}

func (this *cmd) mustTestBinaryByExpectingResponse(t target, expectedPartOfResponse string, args ...string) {
	cmd := append([]string{t.outputName(this.outputPrefix)}, args...)
	response := mustExecuteAndRecord(args...)
	if !strings.Contains(response, expectedPartOfResponse) {
		panic(fmt.Sprintf("Command failed [%s]\nResponse should contain: %s\nBut response was: %s",
			quoteAllIfNeeded(cmd...), expectedPartOfResponse, response))
	}
}
