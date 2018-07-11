package gofn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gofn/gofn/iaas"
	"github.com/gofn/gofn/provision"
	"github.com/nuveo/log"
)

const dockerPort = ":2376"

// ProvideMachine provisioning a machine in the cloud
func ProvideMachine(ctx context.Context, service iaas.Iaas) (client *docker.Client, machine *iaas.Machine, err error) {
	machine, err = service.CreateMachine()
	if err != nil {
		if machine != nil {
			cerr := service.DeleteMachine()
			if cerr != nil {
				log.Errorln(cerr)
			}
		}
		return
	}
	client, err = provision.FnClient(machine.IP+dockerPort, machine.CertsDir)
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
	var machine *iaas.Machine
	done := make(chan struct{})
	go func(ctx context.Context, done chan struct{}) {
		client, err = provision.FnClient("", "")
		if err != nil {
			done <- struct{}{}
			return
		}

		if buildOpts.Iaas != nil {
			client, machine, err = ProvideMachine(ctx, buildOpts.Iaas)
			if err != nil {
				done <- struct{}{}
				return
			}
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
		done <- struct{}{}
	}(ctx, done)
	select {
	case <-ctx.Done():
		log.Errorf("trying to destroy container %v\n", ctx.Err())
	case <-done:
		log.Debugln("trying to destroy container process done")
	}
	if machine != nil {
		log.Debugf("trying to delete machine ID:%v\n", machine.ID)
		deleteErr := buildOpts.Iaas.DeleteMachine()
		if deleteErr != nil {
			err = fmt.Errorf("error trying to delete machine %v", deleteErr)
		}
		return
	}
	if client != nil && container != nil {
		for killAttempt := 0; killAttempt < 3; killAttempt++ {
			if killAttempt > 0 {
				<-time.After(time.Duration(3) * time.Second)
			}
			_, err = provision.FnFindContainerByID(client, container.ID)
			if err != nil {
				if err == provision.ErrContainerNotFound {
					err = nil
				}
				return
			}
			if container.State.Running {
				log.Debugf("destroying container ID:%v, attempt:%v\n", container.ID, killAttempt+1)
				err = client.KillContainer(docker.KillContainerOptions{ID: container.ID})
				if err != nil {
					log.Errorf("error trying to kill container %v, %v, attempt:%v\n", container.ID, err.Error(), killAttempt+1)
				}
			}
			err = client.RemoveContainer(docker.RemoveContainerOptions{
				ID:    container.ID,
				Force: true,
			})
			if err != nil {
				log.Errorf("error trying to remove container %v, %v, attempt:%v\n", container.ID, err.Error(), killAttempt+1)
			}
		}
		err = fmt.Errorf("unable to kill container %v", container.ID)
		return
	}
	err = fmt.Errorf("docker client or container is nil")
	return
}

// DestroyContainer remove by force a container
func DestroyContainer(ctx context.Context, client *docker.Client, container *docker.Container) error {
	return provision.FnRemove(client, container.ID)
}
