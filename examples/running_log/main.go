package main

import (
	"context"
	"log"
	"os"

	"github.com/gofn/gofn"
	"github.com/gofn/gofn/iaas/digitalocean"
	"github.com/gofn/gofn/provision"
)

func main() {
	key := os.Getenv("DIGITALOCEAN_API_KEY")
	if key == "" {
		log.Println("You must provide an api key for digital ocean")
	}
	iaas, err := digitalocean.New(key)
	if err != nil {
		log.Println(err)
	}

	ctx := context.Background()
	client, machine, err := gofn.ProvideMachine(ctx, iaas)
	if err != nil {
		log.Fatal(err)
	}
	defer iaas.DeleteMachine()
	container, err := gofn.PrepareContainer(ctx, client, &provision.BuildOptions{
		ContextDir: "testDocker",
		Dockerfile: "Dockerfile",
		ImageName:  "python",
	}, nil)
	if err != nil {
		log.Println(err)
		return
	}
	errors, err := gofn.RunWait(ctx, client, container)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = gofn.Attach(ctx, client, container, nil, os.Stdout, os.Stderr)
	if err != nil {
		log.Println(err)
		return
	}
	err = <-errors
	if err != nil {
		log.Println(err)
	}
}
