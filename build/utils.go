package main

import (
	"bytes"
	"fmt"
	"github.com/echocat/slf4g"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	versionRegexp = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	startTime     = time.Now()
)

type versionPartAction func(versionPart string)

func executeForVersionParts(version string, action versionPartAction) {
	match := versionRegexp.FindStringSubmatch(version)
	if match != nil {
		action(fmt.Sprintf("v%s.%s", match[1], match[2]))
		action(fmt.Sprintf("v%s", match[1]))
	}
}

func mustExecute(args ...string) {
	mustExecuteTo(nil, os.Stderr, os.Stdout, args...)
}

func mustExecuteAndRecord(args ...string) string {
	buf := new(bytes.Buffer)
	mustExecuteTo(nil, buf, buf, args...)
	return buf.String()
}

type cmdCustomizer func(*exec.Cmd)

func mustExecuteTo(customizer cmdCustomizer, stderr, stdout io.Writer, args ...string) {
	if len(args) <= 0 {
		panic("no arguments provided")
	}
	log.With("command", args).
		Debug("Execute...")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if customizer != nil {
		customizer(cmd)
	}

	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("command failed [%s]: %v", strings.Join(args, " "), err)
		if b, ok := stdout.(fmt.Stringer); ok {
			msg += fmt.Sprintf("\nStdout: %s", b.String())
		}
		if b, ok := stderr.(fmt.Stringer); ok && stderr != stdout {
			msg += fmt.Sprintf("\nStderr: %s", b.String())
		}
		panic(msg)
	}
}
func quoteIfNeeded(what string) string {
	if strings.ContainsRune(what, '\t') ||
		strings.ContainsRune(what, '\n') ||
		strings.ContainsRune(what, ' ') ||
		strings.ContainsRune(what, '\xFF') ||
		strings.ContainsRune(what, '\u0100') ||
		strings.ContainsRune(what, '"') ||
		strings.ContainsRune(what, '\\') {
		return strconv.Quote(what)
	}
	return what
}

func quoteAllIfNeeded(in ...string) []string {
	out := make([]string, len(in))
	for i, a := range in {
		out[i] = quoteIfNeeded(a)
	}
	return out
}
