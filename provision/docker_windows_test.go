package provision

import (
	"testing"
)

func TestFnClient(t *testing.T) {
	type args struct {
		endPoint string
		certsDir string
	}
	tests := []struct {
		name       string
		args       args
		wantClient bool
		wantErr    bool
	}{
		{
			"defaults",
			args{
				"",
				"",
			},
			true,
			false,
		},
		{
			"named pipe",
			args{
				"npipe:////./pipe/docker_engine",
				"",
			},
			true,
			false,
		},
		{
			"error",
			args{
				"http://test:a",
				"",
			},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClient, err := FnClient(tt.args.endPoint, tt.args.certsDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("FnClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotClient != nil) != tt.wantClient {
				t.Errorf("FnClient() = %v, want %v", gotClient, tt.wantClient)
			}
		})
	}
}
