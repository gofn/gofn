package gofn

import (
	docker "github.com/fsouza/go-dockerclient"
	"github.com/nuveo/gofn/provision"
)

// Run runs the designed image
func Run(buildOpts *provision.BuildOptions, volumeOpts *provision.VolumeOptions) (stdout string, err error) {
	client := provision.FnClient("")

	volume := ""
	if volumeOpts != nil {
		volume = provision.FnConfigVolume(volumeOpts)
	}

	img, err := provision.FnFindImage(client, buildOpts.ImageName)
	if err != nil && err != provision.ErrImageNotFound {
		return
	}

	var image string
	var container *docker.Container

	if img.ID == "" {
		image, _ = provision.FnImageBuild(client, buildOpts)
	} else {
		image = "gofn/" + buildOpts.ImageName
	}

	container, err = provision.FnContainer(client, image, volume)
	if err != nil {
		return
	}

	stdout = provision.FnRun(client, container.ID).String()

	err = provision.FnRemove(client, container.ID)
	if err != nil {
		return
	}
	return
}
