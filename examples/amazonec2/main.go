package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofn/gofn"
	"github.com/gofn/gofn/iaas/amazonec2"
	"github.com/gofn/gofn/provision"
)

func main() {
	buildOpts := &provision.BuildOptions{
		ContextDir: "./app",
		Dockerfile: "Dockerfile",
		ImageName:  "gofn-example-1",
		RemoteURI:  "",
		StdIN:      `{"a": 10, "b": 20}`,
	}
	containerOpts := &provision.ContainerOptions{}
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	if accessKey == "" {
		log.Fatalln("You must provide an access key for amazon EC2")
	}
	if secretKey == "" {
		log.Fatalln("You must provide an secret key for amazon EC2")
	}
	p, err := amazonec2.New(accessKey, secretKey)
	if err != nil {
		log.Println(err)
	}
	buildOpts.Iaas = p

	stdout, _, err := gofn.Run(context.Background(), buildOpts, containerOpts)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Stdout: ", stdout)
}
