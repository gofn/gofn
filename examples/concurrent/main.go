package main

import (
	"flag"
	"fmt"
	"log"

	"sync"

	"github.com/nuveo/gofn"
	"github.com/nuveo/gofn/iaas/digitalocean"
	"github.com/nuveo/gofn/provision"
)

const parallels = 3

func main() {
	wait := &sync.WaitGroup{}
	contextDir := flag.String("contextDir", "./", "a string")
	dockerfile := flag.String("dockerfile", "Dockerfile", "a string")
	imageName := flag.String("imageName", "", "a string")
	remoteBuildURI := flag.String("remoteBuildURI", "", "a string")
	volumeSource := flag.String("volumeSource", "", "a string")
	volumeDestination := flag.String("volumeDestination", "", "a string")
	remoteBuild := flag.Bool("remoteBuild", false, "true or false")
	input := flag.String("input", "", "a string")
	flag.Parse()
	wait.Add(parallels)
	for i := 0; i < parallels; i++ {
		run(*contextDir, *dockerfile, *imageName, *remoteBuildURI, *volumeSource, *volumeDestination, wait, *remoteBuild, *input)
	}
	wait.Wait()
}

func run(contextDir, dockerfile, imageName, remoteBuildURI, volumeSource, volumeDestination string, wait *sync.WaitGroup, remote bool, input string) {
	buildOpts := &provision.BuildOptions{
		ContextDir: contextDir,
		Dockerfile: dockerfile,
		ImageName:  imageName,
		RemoteURI:  remoteBuildURI,
		StdIN:      input,
	}
	containerOpts := &provision.ContainerOptions{}
	if volumeSource != "" {
		if volumeDestination == "" {
			volumeDestination = volumeSource
		}
		containerOpts.Volumes = []string{fmt.Sprintf("%s:%s", volumeSource, volumeDestination)}
	}
	if remote {
		buildOpts.Iaas = &digitalocean.Digitalocean{}
	}
	go func() {
		defer wait.Done()
		stdout, stderr, err := gofn.Run(buildOpts, containerOpts)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("Stderr: ", stderr)
		fmt.Println("Stdout: ", stdout)
	}()
}
