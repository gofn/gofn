package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/nuveo/gofn"
	"github.com/nuveo/gofn/provision"
)

func main() {

	contextDir := flag.String("contextDir", "./", "a string")
	dockerfile := flag.String("dockerfile", "Dockerfile", "a string")
	imageName := flag.String("imageName", "", "a string")
	remoteBuildURI := flag.String("remoteBuildURI", "", "a string")
	volumeSource := flag.String("volumeSource", "", "a string")
	volumeDestination := flag.String("volumeDestination", "", "a string")
	flag.Parse()

	stdout, err := gofn.Run(&provision.BuildOptions{
		ContextDir: *contextDir,
		Dockerfile: *dockerfile,
		ImageName:  *imageName,
		RemoteURI:  *remoteBuildURI,
	}, &provision.VolumeOptions{
		Source:      *volumeSource,
		Destination: *volumeDestination,
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println(stdout)
}
