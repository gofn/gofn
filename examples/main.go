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

func main() {
	wait := &sync.WaitGroup{}
	contextDir := flag.String("contextDir", "./", "a string")
	dockerfile := flag.String("dockerfile", "Dockerfile", "a string")
	imageName := flag.String("imageName", "", "a string")
	remoteBuildURI := flag.String("remoteBuildURI", "", "a string")
	volumeSource := flag.String("volumeSource", "", "a string")
	volumeDestination := flag.String("volumeDestination", "", "a string")
	flag.Parse()
	wait.Add(1)
	run(*contextDir, *dockerfile, *imageName, *remoteBuildURI, *volumeSource, *volumeDestination, wait)
	wait.Add(1)
	run(*contextDir, *dockerfile, *imageName, *remoteBuildURI, *volumeSource, *volumeDestination, wait)
	wait.Add(1)
	run(*contextDir, *dockerfile, *imageName, *remoteBuildURI, *volumeSource, *volumeDestination, wait)
	wait.Add(1)
	run(*contextDir, *dockerfile, *imageName, *remoteBuildURI, *volumeSource, *volumeDestination, wait)
	wait.Add(1)
	run(*contextDir, *dockerfile, *imageName, *remoteBuildURI, *volumeSource, *volumeDestination, wait)
	wait.Wait()

}
func run(contextDir, dockerfile, imageName, remoteBuildURI, volumeSource, volumeDestination string, wait *sync.WaitGroup) {
	go func() {
		stdout, err := gofn.Run(&provision.BuildOptions{
			ContextDir: contextDir,
			Dockerfile: dockerfile,
			ImageName:  imageName,
			RemoteURI:  remoteBuildURI,
			Iaas:       &digitalocean.Digitalocean{},
		}, &provision.VolumeOptions{
			Source:      volumeSource,
			Destination: volumeDestination,
		})
		if err != nil {
			log.Println(err)
		}

		fmt.Println(stdout)
		wait.Done()
	}()
}
