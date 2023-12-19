package main

import (
	"runtime"
)

func (this *cmd) mustTest() {
	t := target{os: runtime.GOOS, arch: runtime.GOARCH}
	this.mustTestGoCode(t)
	this.mustBuildBinary(t, true)
	this.mustTestBinary(t)

	if this.withImages {
		this.mustBuildBinary(targetLinuxAmd64, true)
		for _, dv := range imageVariants {
			this.mustBuildImage(dv, true)
			this.testImage(dv)
		}
	}
}
