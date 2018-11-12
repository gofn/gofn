package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofn/gofn"
	"github.com/gofn/gofn/iaas/tcp"
	"github.com/gofn/gofn/provision"
)

func main() {
	// example: docker.gofn.io:2375
	tcp, err := tcp.New("tcp://<your.hosting.com>:2375")
	if err != nil {
		log.Println(err)
		return
	}
	buildOpts := &provision.BuildOptions{
		ContextDir: "./testDocker",
		ImageName:  "tcptesting",
		Iaas:       tcp,
	}
	containerOpts := &provision.ContainerOptions{}
	stdout, stderr, err := gofn.Run(context.Background(), buildOpts, containerOpts)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Stderr: ", stderr)
	fmt.Println("Stdout: ", stdout)
}