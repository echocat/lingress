package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/echocat/lingress/support"
	_ "github.com/echocat/lingress/support/slf4g_native"
	"github.com/echocat/slf4g"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	timeFormat  = "2006-01-02T15:04:05Z"
	basePackage = "github.com/echocat/lingress/support"
)

var (
	targetLinuxAmd64 = target{os: "linux", arch: "amd64"}
	targets          = []target{
		targetLinuxAmd64,
		{os: "linux", arch: "arm64"},
		{os: "windows", arch: "amd64"},
		{os: "windows", arch: "arm64"},
	}
)

func (this *cmd) mustBuildBinaries() {
	for _, t := range targets {
		this.mustBuildBinary(t, false)
	}
}

func (this *cmd) mustBuildBinary(t target, forTesting bool) {
	on := t.outputName(this.outputPrefix)
	support.Must(os.MkdirAll(filepath.Dir(on), 0755))
	started := time.Now()
	l := log.
		With("arch", t.arch).
		With("os", t.os).
		With("output", on).
		With("branch", this.branch).
		With("revision", this.revision).
		With("forTesting", forTesting)
	l.Debug("Build binary...")

	mustExecuteTo(func(cmd *exec.Cmd) {
		cmd.Env = append(os.Environ(), "GOOS="+t.os, "GOARCH="+t.arch, "CGO_ENABLED=0")
	}, os.Stderr, os.Stdout, "go", "build", "-ldflags", this.formatLdFlags(this.branch, this.revision, forTesting), "-o", on, "./main")

	f, err := os.Open(on)
	if err != nil {
		panic(fmt.Errorf("cannot open just created binary %q: %w", on, err))
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		panic(fmt.Errorf("cannot hash just created binary %q: %w", on, err))
	}
	hashHex := hex.EncodeToString(hash.Sum(nil))
	if err := os.WriteFile(on+".sha256", []byte(hashHex), 0644); err != nil {
		panic(fmt.Errorf("cannot safe hash of just created binary %q: %w", on, err))
	}

	l = l.With("duration", time.Now().Sub(started).Truncate(time.Millisecond))
	if forTesting {
		l.Debug("Binary built.")
	} else {
		l.Info("Binary built.")
	}
}

func (this *cmd) formatLdFlags(branch, revision string, forTesting bool) string {
	var testPrefix, testSuffix string
	if forTesting {
		testPrefix = "TEST"
		testSuffix = "TEST"
	}

	ldFlags := "" +
		fmt.Sprintf(" -X %s.branch=%s%s%s", basePackage, testPrefix, branch, testSuffix) +
		fmt.Sprintf(" -X %s.revision=%s%s%s", basePackage, testPrefix, revision, testSuffix) +
		fmt.Sprintf(" -X %s.build=%s", basePackage, time.Now().Format(timeFormat))
	return ldFlags
}

type target struct {
	os   string
	arch string
}

func (this target) outputName(prefix string) string {
	return filepath.Join(fmt.Sprintf("%s-%s-%s%s", prefix, this.os, this.arch, this.ext()))
}

func (this target) ext() string {
	if this.os == "windows" {
		return ".exe"
	}
	return ""
}
