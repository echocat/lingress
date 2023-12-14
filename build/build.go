package main

func (this *cmd) mustBuild() {
	this.mustBuildBinaries()

	if this.withImages {
		this.mustBuildImages()
	}
}
