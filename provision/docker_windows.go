//+build windows

package provision

import (
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
)

// FnClient instantiate a docker client
// For datails https://docs.docker.com/docker-for-windows/faqs/#can-i-use-docker-for-windows-with-new-swarm-mode
// on section "How do I connect to the remote Docker Engine API?"
func FnClient(endPoint, certsDir string) (client *docker.Client, err error) {
	if endPoint == "" {
		endPoint = "npipe:////./pipe/docker_engine"
	}
	if certsDir != "" {
		client, err = docker.NewTLSClient(endPoint, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), filepath.Join(certsDir, "ca.pem"))
		return
	}
	client, err = docker.NewClient(endPoint)
	return
}
