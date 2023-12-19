package main

import (
	log "github.com/echocat/slf4g"
	"time"
)

const imagePrefix = "ghcr.io/echocat/lingress"

var (
	imageVariants = []imageVariant{{
		base: "alpine", file: "Dockerfile", main: true,
	}}
)

func (this *cmd) mustBuildImages() {
	for _, v := range imageVariants {
		this.mustBuildImage(v, false)
		this.mustTagImage(v, false)
	}
}

func (this *cmd) mustBuildImage(v imageVariant, forTesting bool) {
	version := this.branch
	if forTesting {
		version = "TEST" + version + "TEST"
	}

	started := time.Now()
	l := log.
		With("base", v.base).
		With("main", v.main).
		With("file", v.file).
		With("branch", this.branch).
		With("revision", this.revision).
		With("forTesting", forTesting)
	l.Debug("Build image...")

	mustExecute(this.dockerCommand, "build", "-t", v.imageName(version), "-f", v.file, "--build-arg", "image="+imagePrefix, "--build-arg", "version="+version, ".")

	l = l.With("duration", time.Now().Sub(started).Truncate(time.Millisecond))
	if forTesting {
		l.Debug("Image built.")
	} else {
		l.Info("Image built.")
	}
}

func (this *cmd) mustTagImages() {
	for _, v := range imageVariants {
		this.mustTagImage(v, false)
	}
}

func (this *cmd) mustTagImage(v imageVariant, forTesting bool) {
	version := this.branch
	if forTesting {
		version = "TEST" + version + "TEST"
	}
	executeForVersionParts(version, func(tagSuffix string) {
		this.mustTagImageWith(version, v, v.imageName(tagSuffix))
	})
	if this.isLatest {
		this.mustTagImageWith(version, v, v.baseImageName())
	}
	if v.main {
		this.mustTagImageWith(version, v, imagePrefix+":"+version)
		executeForVersionParts(version, func(tagSuffix string) {
			this.mustTagImageWith(version, v, imagePrefix+":"+tagSuffix)
		})
		if this.isLatest {
			this.mustTagImageWith(version, v, imagePrefix+":latest")
		}
	}
}

func (this *cmd) mustTagImageWith(version string, v imageVariant, tag string) {
	mustExecute(this.dockerCommand, "tag", v.imageName(version), tag)
}

type imageVariant struct {
	base string
	file string
	main bool
}

func (this imageVariant) baseImageName() string {
	return imagePrefix + ":" + this.base
}

func (this imageVariant) imageName(branch string) string {
	result := imagePrefix + ":" + this.base
	if branch != "" {
		result += "-" + branch
	}
	return result
}
