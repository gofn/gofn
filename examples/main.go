package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/nuveo/gofn"
)

func main() {

	contextDir := flag.String("contextDir", "./", "a string")
	dockerFile := flag.String("dockerFile", "Dockerfile", "a string")
	imageName := flag.String("imageName", "", "a string")
	flag.Parse()

	stdout, err := gofn.Run(*contextDir, *dockerFile, *imageName)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(stdout)
}
