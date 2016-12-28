package provision

import (
	"bytes"
	"fmt"
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

func FnContainer(client *docker.Client, image string) (Stdout *bytes.Buffer) {
	t := time.Now()
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: fmt.Sprintf("gofn-%s", t.Format("20060102150405")),
		Config: &docker.Config{
			Image:     image,
			StdinOnce: true,
			OpenStdin: true,
		},
	})
	if err != nil {
		panic(err)
	}

	defer client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID, Force: true})
	client.StartContainer(container.ID, nil)
	client.WaitContainerWithContext(container.ID, nil)
	stdout := new(bytes.Buffer)
	client.Logs(docker.LogsOptions{
		Container:    container.ID,
		Stdout:       true,
		OutputStream: stdout,
	})
	Stdout = stdout
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
