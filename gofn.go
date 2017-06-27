package gofn

import (
	"bytes"
	"fmt"
	"io"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nuveo/gofn/iaas"
	"github.com/nuveo/gofn/provision"
)

const dockerPort = ":2375"

// ProvideMachine provisioning a machine in the cloud
func ProvideMachine(service iaas.Iaas) (client *docker.Client, machine *iaas.Machine, err error) {
	machine, err = service.CreateMachine()
	if err != nil {
		if machine != nil {
			cerr := service.DeleteMachine(machine)
			if cerr != nil {
				fmt.Println(cerr.Error())
			}
		}
		return
	}
	client, err = provision.FnClient(machine.IP + dockerPort)
	if err != nil {
		return
	}
	return
}

// PrepareContainer build an image if necessary and run the container
func PrepareContainer(client *docker.Client, buildOpts *provision.BuildOptions, containerOpts *provision.ContainerOptions) (container *docker.Container, err error) {
	img, err := provision.FnFindImage(client, buildOpts.GetImageName())
	if err != nil && err != provision.ErrImageNotFound {
		return
	}

	var image string
	if img.ID == "" {
		image, _, err = provision.FnImageBuild(client, buildOpts)
		if err != nil {
			return
		}
	} else {
		image = buildOpts.GetImageName()
	}

	if containerOpts == nil {
		containerOpts = &provision.ContainerOptions{}
	}
	containerOpts.Image = image
	container, err = provision.FnContainer(client, *containerOpts)
	if err != nil {
		return
	}
	return
}

// RunWait runs the conainer returning channels to control your status
func RunWait(client *docker.Client, container *docker.Container) (errors chan error, err error) {
	err = provision.FnStart(client, container.ID)
	if err != nil {
		return
	}
	errors = provision.FnWaitContainer(client, container.ID)
	return
}

// Attach allow to connect into a running container and interact using stdout, stderr and stdin
func Attach(client *docker.Client, container *docker.Container, stdin io.Reader, stdout io.Writer, stderr io.Writer) (docker.CloseWaiter, error) {
	return provision.FnAttach(client, container.ID, stdin, stdout, stderr)
}

// Run runs the designed image
func Run(buildOpts *provision.BuildOptions, containerOpts *provision.ContainerOptions) (stdout string, stderr string, err error) {
	var client *docker.Client
	client, err = provision.FnClient("")
	if err != nil {
		return
	}

	if buildOpts.Iaas != nil {
		var machine *iaas.Machine
		client, machine, err = ProvideMachine(buildOpts.Iaas)
		if err != nil {
			return
		}
		defer func() {
			err = buildOpts.Iaas.DeleteMachine(machine)
		}()
	}

	var container *docker.Container
	container, err = PrepareContainer(client, buildOpts, containerOpts)
	if err != nil {
		return
	}

	var buffout *bytes.Buffer
	var bufferr *bytes.Buffer

	buffout, bufferr, err = provision.FnRun(client, container.ID, buildOpts.StdIN)
	stdout = buffout.String()
	stderr = bufferr.String()

	DestroyContainer(client, container)
	return
}

// DestroyContainer remove by force a container
func DestroyContainer(client *docker.Client, container *docker.Container) error {
	return provision.FnRemove(client, container.ID)
}
