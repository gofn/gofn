package main

import (
	"encoding/json"
	"fmt"

	"github.com/nuveo/gofn/provision"
)

func main() {
	client := provision.FnClient("")
	imageName, _ := provision.FnImageBuild(client, "testDocker", "Dockerfile", "python")
	stdout := provision.FnContainer(client, imageName)

	if err := json.Unmarshal([]byte(stdout.String()), &provision.DataOut); err != nil {
		panic(err)
	}
	fmt.Println(provision.DataOut)
}
