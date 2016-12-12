package main

import (
	"encoding/json"
	"fmt"

	"github.com/nuveo/gofn/provision"
)

var DataOut map[string]interface{}

func main() {
	client := provision.FnClient("")
	stdout := provision.FnContainer(client, "gofn-python")

	if err := json.Unmarshal([]byte(stdout.String()), &DataOut); err != nil {
		panic(err)
	}
	fmt.Println(DataOut)
}
