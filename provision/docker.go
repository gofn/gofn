package provision

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

func FnClient(endPoint string) (client *docker.Client) {
	if endPoint == "" {
		endPoint = "unix:///var/run/docker.sock"
	}

	client, err := docker.NewClient(endPoint)
	if err != nil {
		panic(err)
	}
	return
}

// FnRemove remove container
func FnRemove(client *docker.Client, containerID string) (err error) {
	err = client.RemoveContainer(docker.RemoveContainerOptions{ID: containerID, Force: true})
	return
}

// FnContainer create container
func FnContainer(client *docker.Client, image string) (container *docker.Container, err error) {
	t := time.Now()
	container, err = client.CreateContainer(docker.CreateContainerOptions{
		Name: fmt.Sprintf("gofn-%s", t.Format("20060102150405")),
		Config: &docker.Config{
			Image:     image,
			StdinOnce: true,
			OpenStdin: true,
		},
	})
	return
}

func FnImageBuild(client *docker.Client, contextDir, dockerFile, imageName string) (Name string, Stdout *bytes.Buffer) {
	if dockerFile == "" {
		dockerFile = "Dockerfile"
	}
	stdout := new(bytes.Buffer)
	Name = "gofn/" + imageName
	err := client.BuildImage(docker.BuildImageOptions{
		Name:           Name,
		Dockerfile:     dockerFile,
		SuppressOutput: true,
		OutputStream:   stdout,
		ContextDir:     contextDir,
	})
	if err != nil {
		panic(err)
	}
	Stdout = stdout
	return
}

// FnFindImage returns image data by name
func FnFindImage(client *docker.Client, imageName string) (image docker.APIImages, err error) {
	var imgs []docker.APIImages
	name := "gofn/" + imageName

	imgs, err = client.ListImages(docker.ListImagesOptions{Filter: name})
	if err != nil {
		return
	}

	if len(imgs) == 0 {
		err = errors.New("Image not found")
		return
	}

	image = imgs[0]
	return
}

// FnFindContainer return container by image name
func FnFindContainer(client *docker.Client, imageName string) (container docker.APIContainers, err error) {
	var containers []docker.APIContainers
	containers, err = client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return
	}

	name := "gofn/" + imageName

	for _, v := range containers {
		if v.Image == name {
			container = v
			return
		}
	}
	err = errors.New("Container not found")
	return
}

func FnKillContainer(client *docker.Client, containerID string) (err error) {
	err = client.KillContainer(docker.KillContainerOptions{ID: containerID})
	return
}

// FnRun runs the container
func FnRun(client *docker.Client, containerID string) (Stdout *bytes.Buffer) {
	err := client.StartContainer(containerID, nil)
	if err != nil {
		log.Println(err)
	}
	client.WaitContainerWithContext(containerID, nil)
	stdout := new(bytes.Buffer)

	client.Logs(docker.LogsOptions{
		Container:    containerID,
		Stdout:       true,
		OutputStream: stdout,
	})
	Stdout = stdout
	return
}
