package provision

import (
	"strings"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	fake "github.com/fsouza/go-dockerclient/testing"
)

func TestFnClientWrongClient(t *testing.T) {
	_, err := FnClient("http://localhost:a")
	if err == nil {
		t.Fatal("expected error but no errors found")
	}
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
	_ = client.PullImage(docker.PullImageOptions{Repository: "gofn/python"}, docker.AuthConfiguration{})
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

	container, err := FnContainer(client, ContainerOptions{Image: image})
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

	_, err := FnContainer(client, ContainerOptions{Image: image})
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
	container, err := FnContainer(client, ContainerOptions{Image: image, Volumes: []string{volume}})
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

func TestFnContainerCreatedWithEnvVars(t *testing.T) {

	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	image := createFakeImage(client)

	env := "GO=fn"
	container, err := FnContainer(client, ContainerOptions{Image: image, Env: []string{env}})
	if err != nil {
		t.Errorf("Expected no errors but %q found", err)
	}
	// container name should starts with gofn
	if !strings.HasPrefix(container.Name, "gofn") {
		t.Errorf("container should starts with gofn but found %q", container.Name)
	}
	// environment variable is set in container
	if container.Config.Env[0] != env {
		t.Errorf("expected  %q bout found %q", env, container.Config.Env[0])
	}

}

func TestFnBuildImageSuccessfully(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	name, _, err := FnImageBuild(client, &BuildOptions{"./testing_data", "", false, "test", "", "", nil})
	if err != nil {
		t.Errorf("FnImageBuild expected nil but found %q, %q", name, err)
	}

	imageName := "gofn/test"
	if name != imageName {
		t.Errorf("image name expected %q but found %q", imageName, name)
	}
}

func TestFnBuildImageRemoteSuccessfully(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	name, _, err := FnImageBuild(client, &BuildOptions{"./testing_data", "", false, "test", "https://github.com/gofn/dockerfile-python-exampl://github.com/gofn/dockerfile-python-example.git", "", nil})
	if err != nil {
		t.Errorf("FnImageBuild expected nil but found %q, %q", name, err)
	}

	imageName := "gofn/test"
	if name != imageName {
		t.Errorf("image name expected %q but found %q", imageName, name)
	}
}

func TestFnBuildImageDoNotUsePrefixImageName(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	imageName := "testDoNotUsePrefixImageName"
	name, _, err := FnImageBuild(client, &BuildOptions{"./testing_data", "", true, imageName, "", "", nil})
	if err != nil {
		t.Errorf("FnImageBuild expected nil but found %q, %q", name, err)
	}
	if name != imageName {
		t.Errorf("image name expected %q but found %q", imageName, name)
	}
}

func TestFnBuildImageDockerfileNotFound(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)
	_, _, err := FnImageBuild(client, &BuildOptions{"./wrong", "Dockerfile", false, "test", "", "", nil})
	if err == nil {
		t.Errorf("FnImageBuild expected error but returned nil")
	}
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

func TestFnFindContainerByIDContainerNotFound(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Find a container by image
	if _, e := FnFindContainerByID(client, "python"); e != ErrContainerNotFound {
		t.Errorf("Expected %q but found %q", ErrContainerNotFound, e)
	}
}

func TestFnListContainers(t *testing.T) {
	// Create a new server
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instanciate the client
	client := NewTestClient(server.URL(), t)

	imageName := createFakeImage(client)

	container, err := FnContainer(client, ContainerOptions{Image: imageName})

	if err != nil {
		t.Fatalf("Expect to create an container but failed because %s", err)
	}

	runFakeContainer(client, container.ID, t)

	containersList, err := FnListContainers(client)

	if err != nil {
		t.Fatalf("Error testing FnListContainers, error: %s", err)
	}

	if len(containersList) == 0 {
		t.Fatal("Expected FnListContainers to have more than zero listed.")
	}
}

func TestFnFindContainerByIDServerError(t *testing.T) {
	client := NewTestClient("wrong", t)

	_, err := FnFindContainerByID(client, "wrong")
	if err == nil || err == ErrContainerNotFound {
		t.Errorf("Expected other errors but found COntainerNotfound or null: %q", err)
	}
}

func TestFnFindContainerByIDSuccessfully(t *testing.T) {
	server := createFakeDockerAPI(t)
	defer server.Stop()

	// Instantiate a client
	client := NewTestClient(server.URL(), t)

	// Create container before find it
	container := createFakeContainer(client, t)

	// Find a container by image
	if _, e := FnFindContainerByID(client, container.ID); e != nil {
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
