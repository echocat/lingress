package main

import (
	"fmt"
	"github.com/echocat/slf4g"
	"strings"
	"time"
)

func (this *cmd) testImage(v imageVariant) {
	started := time.Now()
	l := log.
		With("base", v.base).
		With("main", v.main).
		With("file", v.file)
	l.Debug("Test image...")

	this.mustTestImageByExpectingResponse(v, "Branch:       TEST"+this.branch+"TEST", "--version")
	this.mustTestImageByExpectingResponse(v, "Revision:     TEST"+this.revision+"TEST", "--version")

	l.With("duration", time.Now().Sub(started).Truncate(time.Millisecond)).
		Info("Image tested.")
}

func (this *cmd) mustTestImageByExpectingResponse(v imageVariant, expectedPartOfResponse string, command ...string) {
	call := append([]string{this.dockerCommand, "run", "--rm", v.imageName("TEST" + this.branch + "TEST")}, command...)
	response := mustExecuteAndRecord(call...)
	if !strings.Contains(response, expectedPartOfResponse) {
		panic(fmt.Sprintf("Command failed [%s]\nResponse should contain: %s\nBut response was: %s",
			quoteAllIfNeeded(call...), expectedPartOfResponse, response))
	}
}
