package provision

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gofn/gofn/iaas"
	uuid "github.com/satori/go.uuid"
)

var (
	// ErrImageNotFound is raised when image is not found
	ErrImageNotFound = errors.New("provision: image not found")

	// ErrContainerNotFound is raised when image is not found
	ErrContainerNotFound = errors.New("provision: container not found")

	// ErrContainerExecutionFailed is raised if container exited with status different of zero
	ErrContainerExecutionFailed = errors.New("provision: container exited with failure")

	// Input receives a string that will be written to the stdin of the container in function FnRun
	Input string
)

// BuildOptions are options used in the image build
type BuildOptions struct {
	ContextDir              string
	Dockerfile              string
	DoNotUsePrefixImageName bool
	ImageName               string
	RemoteURI               string
	StdIN                   string
	Iaas                    iaas.Iaas
}

// ContainerOptions are options used in container
type ContainerOptions struct {
	Cmd     []string
	Volumes []string
	Image   string
	Env     []string
}

// GetImageName sets preffix gofn when needed
func (opts BuildOptions) GetImageName() string {
	if opts.DoNotUsePrefixImageName {
		return opts.ImageName
	}
	return "gofn/" + opts.ImageName
}

// FnClient instantiate a docker client
func FnClient(endPoint, certsDir string) (client *docker.Client, err error) {
	if endPoint == "" {
		endPoint = "unix:///var/run/docker.sock"
	}

	client, err = docker.NewTLSClient(endPoint, certsDir+"/cert.pem", certsDir+"/key.pem", certsDir+"/ca.pem")
	return
}

// FnRemove remove container
func FnRemove(client *docker.Client, containerID string) (err error) {
	err = client.RemoveContainer(docker.RemoveContainerOptions{ID: containerID, Force: true})
	return
}

// FnContainer create container
func FnContainer(client *docker.Client, opts ContainerOptions) (container *docker.Container, err error) {
	config := &docker.Config{
		Image:     opts.Image,
		Cmd:       opts.Cmd,
		Env:       opts.Env,
		StdinOnce: true,
		OpenStdin: true,
	}
	var uid uuid.UUID
	uid, err = uuid.NewV4()
	if err != nil {
		return
	}
	container, err = client.CreateContainer(docker.CreateContainerOptions{
		Name:       fmt.Sprintf("gofn-%s", uid.String()),
		HostConfig: &docker.HostConfig{Binds: opts.Volumes},
		Config:     config,
	})
	return
}

// FnImageBuild builds an image
func FnImageBuild(client *docker.Client, opts *BuildOptions) (Name string, Stdout *bytes.Buffer, err error) {
	if opts.Dockerfile == "" {
		opts.Dockerfile = "Dockerfile"
	}
	stdout := new(bytes.Buffer)
	Name = opts.GetImageName()
	err = client.BuildImage(docker.BuildImageOptions{
		Name:           Name,
		Dockerfile:     opts.Dockerfile,
		SuppressOutput: true,
		OutputStream:   stdout,
		ContextDir:     opts.ContextDir,
		Remote:         opts.RemoteURI,
	})
	if err != nil {
		return
	}
	Stdout = stdout
	return
}

// FnFindImage returns image data by name
func FnFindImage(client *docker.Client, imageName string) (image docker.APIImages, err error) {
	var imgs []docker.APIImages
	imgs, err = client.ListImages(docker.ListImagesOptions{Filter: imageName})
	if err != nil {
		return
	}
	if len(imgs) == 0 {
		err = ErrImageNotFound
		return
	}
	image = imgs[0]
	return
}

// FnFindContainerByID return container by ID
func FnFindContainerByID(client *docker.Client, ID string) (container docker.APIContainers, err error) {
	var containers []docker.APIContainers
	containers, err = client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return
	}
	for _, v := range containers {
		if v.ID == ID {
			container = v
			return
		}
	}
	err = ErrContainerNotFound
	return
}

// FnFindContainer return container by image name
func FnFindContainer(client *docker.Client, imageName string) (container docker.APIContainers, err error) {
	var containers []docker.APIContainers
	containers, err = client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return
	}

	if !strings.HasPrefix(imageName, "gofn") {
		imageName = "gofn/" + imageName
	}

	for _, v := range containers {
		if v.Image == imageName {
			container = v
			return
		}
	}
	err = ErrContainerNotFound
	return
}

// FnKillContainer kill the container
func FnKillContainer(client *docker.Client, containerID string) (err error) {
	err = client.KillContainer(docker.KillContainerOptions{ID: containerID})
	return
}

//FnAttach attach into a running container
func FnAttach(client *docker.Client, containerID string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (w docker.CloseWaiter, err error) {
	return client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    containerID,
		RawTerminal:  true,
		Stream:       true,
		Stdin:        true,
		Stderr:       true,
		Stdout:       true,
		Logs:         true,
		InputStream:  stdin,
		ErrorStream:  stderr,
		OutputStream: stdout,
	})
}

// FnStart start the container
func FnStart(client *docker.Client, containerID string) error {
	return client.StartContainer(containerID, nil)
}

// FnRun runs the container
func FnRun(client *docker.Client, containerID, input string) (Stdout *bytes.Buffer, Stderr *bytes.Buffer, err error) {
	err = FnStart(client, containerID)
	if err != nil {
		return
	}

	// attach to write input
	_, err = FnAttach(client, containerID, strings.NewReader(input), nil, nil)
	if err != nil {
		return
	}

	e := FnWaitContainer(client, containerID)
	err = <-e

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// omit logs because execution error is more important
	_ = FnLogs(client, containerID, stdout, stderr)

	Stdout = stdout
	Stderr = stderr
	return
}

// FnLogs logs all container activity
func FnLogs(client *docker.Client, containerID string, stdout io.Writer, stderr io.Writer) error {
	return client.Logs(docker.LogsOptions{
		Container:    containerID,
		Stdout:       true,
		Stderr:       true,
		ErrorStream:  stderr,
		OutputStream: stdout,
	})
}

// FnWaitContainer wait until container finnish your processing
func FnWaitContainer(client *docker.Client, containerID string) chan error {
	errs := make(chan error)
	go func() {
		code, err := client.WaitContainer(containerID)
		if err != nil {
			errs <- err
		}
		if code != 0 {
			errs <- ErrContainerExecutionFailed
		}
		errs <- nil
	}()
	return errs
}

// FnListContainers lists all the containers created by the gofn.
// It returns the APIContainers from the API, but have to be formatted for pretty printing
func FnListContainers(client *docker.Client) (containers []docker.APIContainers, err error) {
	hostContainers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
	})
	if err != nil {
		containers = nil
		return
	}
	for _, container := range hostContainers {
		if strings.HasPrefix(container.Image, "gofn/") {
			containers = append(containers, container)
		}
	}
	return
}
