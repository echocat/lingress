package main

import (
	log "github.com/echocat/slf4g"
	"time"
)

func (this *cmd) mustDeployImages() {
	for _, v := range imageVariants {
		this.mustDeployImage(v)
	}
}

func (this *cmd) mustDeployImage(v imageVariant) {
	started := time.Now()
	l := log.
		With("base", v.base).
		With("main", v.main).
		With("file", v.file).
		With("branch", this.branch).
		With("revision", this.revision)
	l.Debug("Deploy image...")

	this.mustDeployImageTag(v.imageName(this.branch))
	executeForVersionParts(this.branch, func(tagSuffix string) {
		this.mustDeployImageTag(v.imageName(tagSuffix))
	})
	if this.isLatest {
		this.mustDeployImageTag(v.baseImageName())
	}
	if v.main {
		this.mustDeployImageTag(imagePrefix + ":" + this.branch)
		executeForVersionParts(this.branch, func(tagSuffix string) {
			this.mustDeployImageTag(imagePrefix + ":" + tagSuffix)
		})
		if this.isLatest {
			this.mustDeployImageTag(imagePrefix + ":latest")
		}
	}

	l.With("duration", time.Now().Sub(started).Truncate(time.Millisecond)).
		Info("Image deployed.")
}

func (this *cmd) mustDeployImageTag(tag string) {
	mustExecute(this.dockerCommand, "push", tag)
}
