package main

import (
	"encoding/json"
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nuveo/gofn/provision"
)

func main() {
	client := provision.FnClient("")

	img, err := provision.FnFindImage(client, "python")
	if err != nil {
		panic(err)
	}

	var imageName string
	var container *docker.Container

	if img.ID == "" {
		imageName, _ = provision.FnImageBuild(client, "testDocker", "Dockerfile", "python")
	} else {
		imageName = "gofn/" + "python"
	}

	container, err = provision.FnContainer(client, imageName)
	if err != nil {
		panic(err)
	}

	stdout := provision.FnRun(client, container.ID)

	provision.FnRemove(client, container.ID)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal([]byte(stdout.String()), &provision.DataOut); err != nil {
		panic(err)
	}
	fmt.Println(provision.DataOut)
}
