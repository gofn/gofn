package provision

import (
	"testing"

	docker "github.com/fsouza/go-dockerclient"
)

func TestFnClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	var client *docker.Client
	// connect from local socket
	client = FnClient("unix:///var/run/docker.sock")
	if err := client.Ping(); err != nil {
		t.Errorf("docker.sock: expected nil but found %q", err)
	}

	// Empty string also connect local socket
	client = FnClient("")
	if err := client.Ping(); err != nil {
		t.Errorf("empty string: expected nil but found %q", err)
	}
}

// maybe setup with containers and images
