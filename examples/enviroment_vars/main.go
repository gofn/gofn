package main

import (
	"fmt"
	"log"

	"github.com/nuveo/gofn"
	"github.com/nuveo/gofn/provision"
)

func main() {
	buildOpts := &provision.BuildOptions{
		ContextDir: "./testDocker",
		ImageName:  "gofn/env",
	}
	containerOpts := &provision.ContainerOptions{
		Env: []string{"FOO=bar", "KEY=value"},
		Cmd: []string{"env"},
	}
	stdout, stderr, err := gofn.Run(buildOpts, containerOpts)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Stderr: ", stderr)
	fmt.Println("Stdout: ", stdout)
}
