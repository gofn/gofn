package gofn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nuveo/gofn/iaas"
	"github.com/nuveo/gofn/provision"
)

const dockerPort = ":2375"

// ProvideMachine provisioning a machine in the cloud
func ProvideMachine(ctx context.Context, service iaas.Iaas) (client *docker.Client, machine *iaas.Machine, err error) {
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
func PrepareContainer(ctx context.Context, client *docker.Client, buildOpts *provision.BuildOptions, containerOpts *provision.ContainerOptions) (container *docker.Container, err error) {
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
func RunWait(ctx context.Context, client *docker.Client, container *docker.Container) (errors chan error, err error) {
	err = provision.FnStart(client, container.ID)
	if err != nil {
		return
	}
	errors = provision.FnWaitContainer(client, container.ID)
	return
}

// Attach allow to connect into a running container and interact using stdout, stderr and stdin
func Attach(ctx context.Context, client *docker.Client, container *docker.Container, stdin io.Reader, stdout io.Writer, stderr io.Writer) (docker.CloseWaiter, error) {
	return provision.FnAttach(client, container.ID, stdin, stdout, stderr)
}

// Run runs the designed image
func Run(ctx context.Context, buildOpts *provision.BuildOptions, containerOpts *provision.ContainerOptions) (stdout string, stderr string, err error) {
	var client *docker.Client
	var container *docker.Container
	done := make(chan struct{})
	go func(ctx context.Context, done chan struct{}) {
		client, err = provision.FnClient("")
		if err != nil {
			done <- struct{}{}
			return
		}

		if buildOpts.Iaas != nil {
			var machine *iaas.Machine
			client, machine, err = ProvideMachine(ctx, buildOpts.Iaas)
			if err != nil {
				done <- struct{}{}
				return
			}
			defer func() {
				err = buildOpts.Iaas.DeleteMachine(machine)
			}()
		}

		container, err = PrepareContainer(ctx, client, buildOpts, containerOpts)
		if err != nil {
			done <- struct{}{}
			return
		}

		var buffout *bytes.Buffer
		var bufferr *bytes.Buffer

		buffout, bufferr, err = provision.FnRun(client, container.ID, buildOpts.StdIN)
		stdout = buffout.String()
		stderr = bufferr.String()

	}(ctx, done)
	select {
	case <-ctx.Done():
		fmt.Printf("trying to destroy container %v\n", ctx.Err())
	case <-done:
		fmt.Println("trying to destroy container process done")
	}
	if client != nil && container != nil {
		derr := DestroyContainer(context.Background(), client, container)
		if derr != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "docker client and container is nil")
	return
}

// DestroyContainer remove by force a container
func DestroyContainer(ctx context.Context, client *docker.Client, container *docker.Container) error {
	return provision.FnRemove(client, container.ID)
}
