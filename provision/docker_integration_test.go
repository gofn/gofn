package provision

import "testing"

func TestFnClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// connect from local socket
	client, err := FnClient("unix:///var/run/docker.sock", "")
	if err != nil {
		t.Errorf("FnClient with URI: expected nil but returned %q", err)
	}

	err = client.Ping()
	if err != nil {
		t.Errorf("docker.sock: expected nil but found %q", err)
	}

	// Empty string also connect local socket
	client, err = FnClient("", "")
	if err != nil {
		t.Errorf("FnClient: expected nil but returned %q", err)
	}

	if err = client.Ping(); err != nil {
		t.Errorf("empty string: expected nil but found %q", err)
	}
}

// maybe setup with containers and images
