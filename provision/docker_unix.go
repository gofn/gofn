//+build !windows

package provision

import (
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
)

// FnClient instantiate a docker client
func FnClient(endPoint, certsDir string) (client *docker.Client, err error) {
	if endPoint == "" {
		endPoint = "unix:///var/run/docker.sock"
	}
	if certsDir != "" {
		client, err = docker.NewTLSClient(endPoint, filepath.Join(certsDir, "cert.pem"), filepath.Join(certsDir, "key.pem"), filepath.Join(certsDir, "ca.pem"))
		return
	}
	client, err = docker.NewClient(endPoint)
	return
}
