package gofn

import (
	"bytes"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nuveo/gofn/iaas"
	"github.com/nuveo/gofn/provision"
)

const dockerPort = ":2375"

// Input receives a string that will be written to the stdin of the container
var Input string

// Run runs the designed image
func Run(buildOpts *provision.BuildOptions, volumeOpts *provision.VolumeOptions) (stdout string, stderr string, err error) {
	var client *docker.Client
	client, err = provision.FnClient("")
	if err != nil {
		return
	}

	if buildOpts.Iaas != nil {
		var machine *iaas.Machine
		machine, err = buildOpts.Iaas.CreateMachine()
		if err != nil {
			if machine != nil {
				err = buildOpts.Iaas.DeleteMachine(machine)
			}
			return
		}
		defer buildOpts.Iaas.DeleteMachine(machine)
		client, err = provision.FnClient(machine.IP + dockerPort)
		if err != nil {
			return
		}
	}

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
		image, _, err = provision.FnImageBuild(client, buildOpts)
		if err != nil {
			return
		}
	} else {
		image = "gofn/" + buildOpts.ImageName
	}

	container, err = provision.FnContainer(client, image, volume)
	if err != nil {
		return
	}

	var buffout *bytes.Buffer
	var bufferr *bytes.Buffer

	provision.Input = Input
	buffout, bufferr, err = provision.FnRun(client, container.ID)
	if err != nil {
		return
	}
	stdout = buffout.String()
	stderr = bufferr.String()

	err = provision.FnRemove(client, container.ID)
	if err != nil {
		return
	}
	return
}
