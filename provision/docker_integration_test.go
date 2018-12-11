package provision

import (
	"os"
	"strings"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func TestFnImageBuildIntegration(t *testing.T) {
	client, err := FnClient("", "")
	if err != nil {
		t.Fatal(err)
	}

	opts := &BuildOptions{
		ImageName: "nuveo/testprivategofn",
		Auth: docker.AuthConfiguration{
			Username: os.Getenv("DOCKER_LOGIN"),
			Password: os.Getenv("DOCKER_PASSWORD"),
		},
	}
	name, out, err := FnImageBuild(client, opts)
	if err != nil {
		t.Fatal(err)
	}
	if name != "nuveo/testprivategofn" {
		t.Error(`image name is not "nuveo/testprivategofn"`)
	}
	if !strings.Contains(out.String(), "Pull") {
		t.Errorf(`do not contains word "Pull", OUT: %v`, out.String())
	}
}
