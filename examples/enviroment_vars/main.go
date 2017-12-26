package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofn/gofn"
	"github.com/gofn/gofn/provision"
)

func main() {
	buildOpts := &provision.BuildOptions{
		ContextDir: "./testDocker",
		ImageName:  "env",
	}
	containerOpts := &provision.ContainerOptions{
		Env: []string{"FOO=bar", "KEY=value"},
		Cmd: []string{"env"},
	}
	stdout, stderr, err := gofn.Run(context.Background(), buildOpts, containerOpts)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Stderr: ", stderr)
	fmt.Println("Stdout: ", stdout)
}
