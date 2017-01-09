package provision

import (
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

func createFakeDockerAPI(t *testing.T) *fake.DockerServer {
	// Fake docker api
	server, err := fake.NewServer("127.0.0.1:0", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func createFakeImage(client *docker.Client) string {
	client.PullImage(docker.PullImageOptions{Repository: "gofn/python"}, docker.AuthConfiguration{})
	return "gofn/python"
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

func runFakeContainer(client *docker.Client, containerID string, t *testing.T) {
	err := client.StartContainer(containerID, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func NewTestClient(host string, t *testing.T) *docker.Client {
	client, err := docker.NewClient(host)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestFnRemoveContainerSuccessfully(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Create container before remove it
	container := createFakeContainer(client, t)

	// Remove a container
	if e := FnRemove(client, container.ID); e != nil {
		t.Errorf("Expected no errors but %q found", e)
	}
}

func TestFnRemoveContainerImageNotFound(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Remove a container
	if e := FnRemove(client, "wrong"); e == nil {
		t.Error("expected errors but no errors found")
	}
}

func TestFnContainerCreatedSuccessfully(t *testing.T) {

	server := createFakeDockerAPI(t)
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

}

func TestFnContainerInvalidImage(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	image := "wrong"

	_, err := FnContainer(client, image, "")
	if err == nil {
		t.Errorf("Expected errors but no errors found")
	}

}

func TestFnContainerCreatedWithVolume(t *testing.T) {

	server := createFakeDockerAPI(t)
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
}
func TestFnBuildImageSuccessfully(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	name, _ := FnImageBuild(client, "./testing_data", "", "test")

	imageName := "gofn/test"
	if name != imageName {
		t.Errorf("image name expected %q but found %q", imageName, name)
	}
}

func TestFnBuildImageDockerfileNotFound(t *testing.T) {
	server := createFakeDockerAPI(t)
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
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	img := createFakeImage(client)

	_, err := FnFindImage(client, img)
	if err != nil {
		t.Errorf("no errors expected but found %q", err)
	}
}

func TestFnFindImageImageNotFound(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	_, err := FnFindImage(client, "wrong")
	if err != ErrImageNotFound {
		t.Errorf("unexpected error: %q", err)
	}
}

func TestFnFindImageServerError(t *testing.T) {
	client := NewTestClient("wrong", t)

	_, err := FnFindImage(client, "wrong")
	if err == nil || err == ErrImageNotFound {
		t.Errorf("Expected other errors but found ImageNotfound or null: %q", err)
	}
}

func TestFnFindContainerSuccessfully(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Create container before find it
	container := createFakeContainer(client, t)

	// Find a container by image
	if _, e := FnFindContainer(client, container.Image); e != nil {
		t.Errorf("Expected no errors but %q found", e)
	}
}

func TestFnFindContainerContainerNotFound(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Find a container by image
	if _, e := FnFindContainer(client, "python"); e != ErrContainerNotFound {
		t.Errorf("Expected %q but found %q", ErrContainerNotFound, e)
	}
}

func TestFnFindContainerServerError(t *testing.T) {
	client := NewTestClient("wrong", t)

	_, err := FnFindContainer(client, "wrong")
	if err == nil || err == ErrContainerNotFound {
		t.Errorf("Expected other errors but found COntainerNotfound or null: %q", err)
	}
}

func TestFnKilContainerSuccessfully(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()
	// instantiate a client
	client := NewTestClient(server.URL(), t)

	// create container before kill it
	container := createFakeContainer(client, t)

	// make sure that container are runnig
	runFakeContainer(client, container.ID, t)

	// kill a container
	if e := FnKillContainer(client, container.ID); e != nil {
		t.Errorf("Expected no errors but %q found", e)
	}

}

func TestFnKilContainerNotRunning(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Create container before kill it
	container := createFakeContainer(client, t)

	// kill a container
	if e := FnKillContainer(client, container.ID); e == nil {
		t.Errorf("expecting errors, but nothing found")
	}
}

func TestFnKilContainerNotFound(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// kill a container
	fakeID := "wrong123"
	if e := FnKillContainer(client, fakeID); e == nil {
		t.Errorf("expecting errors, but nothing found")
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
