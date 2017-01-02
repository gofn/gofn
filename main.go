package main

import (
	"encoding/json"
	"flag"
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/nuveo/gofn/provision"
)

func main() {

	contextDir := flag.String("contextDir", "./", "a string")
	dockerFile := flag.String("dockerFile", "Dockerfile", "a string")
	imageName := flag.String("imageName", "", "a string")
	flag.Parse()

	client := provision.FnClient("")

	img, err := provision.FnFindImage(client, "python")
	if err != nil {
		panic(err)
	}

	var image string
	var container *docker.Container

	if img.ID == "" {
		image, _ = provision.FnImageBuild(client, *contextDir, *dockerFile, *imageName)
	} else {
		image = "gofn/" + (*imageName)
	}

	container, err = provision.FnContainer(client, image)
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
