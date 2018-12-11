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
		ImageName:               "nuveo/testprivategofn",
		DoNotUsePrefixImageName: true,
		ContextDir:              "./",
		Auth: docker.AuthConfiguration{
			Username: os.Getenv("DOCKER_LOGIN"),
			Password: os.Getenv("DOCKER_PASSWORD"),
		},
	}
	name, _, err := FnImageBuild(client, opts)
	if err != nil {
		t.Fatal(err)
	}
	if name != "nuveo/testprivategofn" {
		t.Error(`image name is not "nuveo/testprivategofn"`)
	}
	c, err := FnContainer(client, ContainerOptions{
		Image: opts.ImageName,
	})
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := FnRun(client, c.ID, "test")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Error("stderr is not empty")
	}
	if strings.TrimSpace(stdout.String()) == "" {
		t.Error("stdout is empty")
	}
}
