package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/fsouza/go-dockerclient"
)

var Data map[string]interface{}

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: "gofn",
		Config: &docker.Config{
			Image:     "gofn-python",
			StdinOnce: true,
			OpenStdin: true,
		},
	})
	if err != nil {
		panic(err)
	}
	defer client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID, Force: true})

	// Star container
	client.StartContainer(container.ID, nil)

	stdout := new(bytes.Buffer)
	client.Logs(docker.LogsOptions{
		Container:    container.ID,
		Stdout:       true,
		OutputStream: stdout,
	})

	if err := json.Unmarshal([]byte(stdout.String()), &Data); err != nil {
		panic(err)
	}
	fmt.Println(Data)
}
