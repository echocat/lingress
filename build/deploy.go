package main

func (this *cmd) mustDeploy() {
	if this.withImages {
		this.mustDeployImages()
	}
}
