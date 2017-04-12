package main

import (
	"log"
	"os"

	"github.com/nuveo/gofn"
	"github.com/nuveo/gofn/iaas/digitalocean"
	"github.com/nuveo/gofn/provision"
)

func main() {
	iaas := &digitalocean.Digitalocean{}
	client, machine, err := gofn.ProvideMachine(iaas)
	if err != nil {
		log.Fatal(err)
	}
	defer iaas.DeleteMachine(machine)
	container, err := gofn.PrepareContainer(client, &provision.BuildOptions{
		ContextDir: "testDocker",
		Dockerfile: "Dockerfile",
		ImageName:  "python",
	}, nil)
	if err != nil {
		log.Println(err)
		return
	}
	errors, err := gofn.RunWait(client, container)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = gofn.Attach(client, container, nil, os.Stdout, os.Stderr)
	if err != nil {
		log.Println(err)
		return
	}
	err = <-errors
	if err != nil {
		log.Println(err)
	}
}
