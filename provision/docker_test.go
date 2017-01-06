package provision

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	fake "github.com/fsouza/go-dockerclient/testing"
)

func TestFnClientWrongClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	FnClient("http://localhost:a")
}

func createFakeDockerAPI(requests *[]*http.Request, t *testing.T) *fake.DockerServer {
	// Fake docker api
	server, err := fake.NewServer("127.0.0.1:0", nil, func(r *http.Request) {
		*requests = append(*requests, r)
	})
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func createFakeImage(client *docker.Client) string {
	client.PullImage(docker.PullImageOptions{Repository: "python"}, docker.AuthConfiguration{})
	return "python"
}

func createFakeContainer(client *docker.Client, t *testing.T) *docker.Container {
	image := createFakeImage(client)
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: "gofn-123",
		Config: &docker.Config{
			Image:     image,
			StdinOnce: true,
			OpenStdin: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return container
}

func NewTestClient(host string, t *testing.T) *docker.Client {
	client, err := docker.NewClient(host)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestFnRemoveContainerSuccessfully(t *testing.T) {

	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Create container before remove it
	container := createFakeContainer(client, t)

	// Remove a container
	if e := FnRemove(client, container.ID); e != nil {
		t.Errorf("Expected no errors but %q found", e)
	}

	// last request
	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodDelete {
		t.Errorf("expected method DELETE but %q found", lastRequest.Method)
	}
	if lastRequest.URL.Path != fmt.Sprintf("/containers/%s", container.ID) {
		t.Errorf("expected \"containers/%s\" but path was %q", container.ID, lastRequest.URL.Path)
	}

}

func TestFnRemoveContainerImageNotFound(t *testing.T) {

	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Remove a container
	if e := FnRemove(client, "wrong"); e == nil {
		t.Error("expected errors but no errors found")
	}

	// last request
	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodDelete {
		t.Errorf("expected method DELETE but %q found", lastRequest.Method)
	}
	if lastRequest.URL.Path != fmt.Sprintf("/containers/%s", "wrong") {
		t.Errorf("expected \"containers/%s\" but path was %q", "wrong", lastRequest.URL.Path)
	}
}

func TestFnContainerCreatedSuccessfully(t *testing.T) {

	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	image := createFakeImage(client)

	container, err := FnContainer(client, image, "")
	if err != nil {
		t.Errorf("Expected no errors but %q found", err)
	}
	// container name should starts with gofn
	if !strings.HasPrefix(container.Name, "gofn") {
		t.Errorf("container should starts with gofn but found %q", container.Name)
	}
	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodPost {
		t.Errorf("expected method %q but %q found", http.MethodPost, lastRequest.Method)
	}
	if !strings.Contains(lastRequest.URL.Path, "/containers/create") {
		t.Errorf("Path is not ok, found: %q", lastRequest.URL.Path)
	}

}

func TestFnContainerInvalidImage(t *testing.T) {

	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	image := "wrong"

	_, err := FnContainer(client, image, "")
	if err == nil {
		t.Errorf("Expected errors but no errors found")
	}

	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodPost {
		t.Errorf("expected method %q but %q found", http.MethodPost, lastRequest.Method)
	}
	if !strings.Contains(lastRequest.URL.Path, "/containers/create") {
		t.Errorf("Path is not ok, found: %q", lastRequest.URL.Path)
	}

}

func TestFnContainerCreatedWithVolume(t *testing.T) {

	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	image := createFakeImage(client)

	volume := "/tmp:/tmp"
	container, err := FnContainer(client, image, volume)
	if err != nil {
		t.Errorf("Expected no errors but %q found", err)
	}
	// container name should starts with gofn
	if !strings.HasPrefix(container.Name, "gofn") {
		t.Errorf("container should starts with gofn but found %q", container.Name)
	}
	// volume is binded with container
	if container.HostConfig.Binds[0] != volume {
		t.Errorf("expected volume %q bout found %q", volume, container.HostConfig.Binds[0])
	}
	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodPost {
		t.Errorf("expected method %q but %q found", http.MethodPost, lastRequest.Method)
	}
	if !strings.Contains(lastRequest.URL.Path, "/containers/create") {
		t.Errorf("Path is not ok, found: %q", lastRequest.URL.Path)
	}

}
func TestFnBuildImageSuccessfully(t *testing.T) {
	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	name, _ := FnImageBuild(client, "./testing_data", "", "test")

	imageName := "gofn/test"
	if name != imageName {
		t.Errorf("image name expected %q but found %q", imageName, name)
	}

	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodPost {
		t.Errorf("expected method %q but %q found", http.MethodPost, lastRequest.Method)
	}
	if !strings.Contains(lastRequest.URL.Path, "/build") {
		t.Errorf("Path is not ok, found: %q", lastRequest.URL.Path)
	}
}

func TestFnBuildImageDockerfileNotFound(t *testing.T) {
	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()
	FnImageBuild(client, "./wrong", "Dockerfile", "test")
}

func TestFnFindImageSuccessfully(t *testing.T) {
	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	img := createFakeImage(client)

	_, err := FnFindImage(client, img)
	if err != nil {
		t.Errorf("no errors expected but found %q", err)
	}
	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodGet {
		t.Errorf("expected method %q but %q found", http.MethodGet, lastRequest.Method)
	}
	if !strings.Contains(lastRequest.URL.Path, "/images/json") {
		t.Errorf("Path is not ok, found: %q", lastRequest.URL.Path)
	}
}

func TestFnFindImageImageNotFound(t *testing.T) {
	var requests []*http.Request
	server := createFakeDockerAPI(&requests, t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	_, err := FnFindImage(client, "wrong")
	if err != ErrImageNotFound {
		t.Errorf("unexpected error: %q", err)
	}
	lastRequest := requests[len(requests)-1]
	if lastRequest.Method != http.MethodGet {
		t.Errorf("expected method %q but %q found", http.MethodGet, lastRequest.Method)
	}
	if !strings.Contains(lastRequest.URL.Path, "/images/json") {
		t.Errorf("Path is not ok, found: %q", lastRequest.URL.Path)
	}
}

func TestFnFindImageServerError(t *testing.T) {
	client := NewTestClient("wrong", t)

	_, err := FnFindImage(client, "wrong")
	if err == nil || err == ErrImageNotFound {
		t.Errorf("Expected other errors but found ImageNotfound or null: %q", err)
	}
}

func TestFnConfigVolumeAllEmpty(t *testing.T) {
	volume := FnConfigVolume(&VolumeOptions{})
	if volume != "" {
		t.Errorf("Expected \"\" but found %q", volume)
	}
}

func TestFnConfigVolumeDestinationOmitted(t *testing.T) {
	volume := FnConfigVolume(&VolumeOptions{Source: "/tmp"})
	if volume != "/tmp:/tmp" {
		t.Errorf("Expected \"/tmp:/tmp\" but found %q", volume)
	}
}

func TestFnConfigVolumeOnlyDestination(t *testing.T) {
	volume := FnConfigVolume(&VolumeOptions{Destination: "/tmp"})
	if volume != ":/tmp" {
		t.Errorf("Expected \":/tmp\" but found %q", volume)
	}
}

func TestFnConfigVolumeAllFields(t *testing.T) {
	volume := FnConfigVolume(&VolumeOptions{
		Source:      "/tmp",
		Destination: "/tmp",
	})
	if volume != "/tmp:/tmp" {
		t.Errorf("Expected \"/tmp:/tmp\" but found %q", volume)
	}
}
