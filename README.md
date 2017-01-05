# gofn

[![Build Status](https://travis-ci.org/nuveo/gofn.svg?branch=master)](https://travis-ci.org/nuveo/gofn)
[![GoDoc](https://godoc.org/github.com/nuveo/gofn?status.png)](https://godoc.org/github.com/nuveo/gofn)
[![Go Report Card](https://goreportcard.com/badge/github.com/nuveo/gofn)](https://goreportcard.com/report/github.com/nuveo/gofn)

Function process via container provider

## Premise

- Software makes a task
- After processing it dies
- Must put on stdout (print) a string formatted as JSON

## Install

```bash
go get github.com/nuveo/gofn
```

## Example

```go
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
	dockerFile := flag.String("dockerFile", "Dockerfile", "a string")
	imageName := flag.String("imageName", "", "a string")
	volumeSource := flag.String("volumeSource", "", "a string")
	volumeDestination := flag.String("volumeDestination", "", "a string")
	flag.Parse()

	stdout, err := gofn.Run(*contextDir, *dockerFile, *imageName, &provision.VolumeOptions{
		Source:      *volumeSource,
		Destination: *volumeDestination,
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println(stdout)
}
```

### Run Example

```bash
	cd examples
	go run main.go -contextDir=testDocker -imageName=python -dockerFile=Dockerfile
	# or using volume
	go run main.go -contextDir=testDocker -imageName=python -dockerFile=Dockerfile -volumeSource=/tmp -volumeDestination=/tmp
```

You can also compile with _go build_ or build and install with _go install_ command then run it as a native executable.

### Example Parameters

- -contextDir is the root directory where the Dockerfile, scripts, and other container dependencies are, by default current directory "./".

- -imageName is the name of the image you want to start, if it does not exist it will be automatically generated and if it exists the system will just start the container.

- -dockerFile is the name of the file containing the container settings, by default Dockerfile

- volumeSource is the directory that will be mounted as a data volume. By default is empty string indicating his not used.

- volumeDestination is the path mounted inside the container. By default is empty string indicating his  not used but if only omitted, volumeSource is used.

- -h Shows the list of parameters

gofn generates the images with "gofn/" as a prefix.
