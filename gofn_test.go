package gofn

import (
	"context"
	"testing"

	"github.com/nuveo/gofn/provision"
)

func TestRun(t *testing.T) {

	buildOpts := &provision.BuildOptions{
		ContextDir: "./error_path", // this path must not exist
		ImageName:  "testgofn",
	}
	_, _, err := Run(context.Background(), buildOpts, nil)
	if err == nil {
		t.Fatal("Expected error but returned nil, this test must fail because the path to Dockerfile does not exist")
	}

}
