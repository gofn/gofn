package tcp

import (
	"reflect"
	"testing"

	"github.com/gofn/gofn/iaas"
)

func TestNew(t *testing.T) {
	type args struct {
		URL string
	}
	tests := []struct {
		name    string
		args    args
		wantP   *Provider
		wantErr bool
	}{
		{
			"success",
			args{
				"tcp://localhost:2376",
			},
			&Provider{
				Host: "localhost",
				Port: 2376,
			},
			false,
		},
		{
			"parse error",
			args{
				"a",
			},
			nil,
			true,
		},
		{
			"not tcp error",
			args{
				"http://localhost:2376",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP, err := New(tt.args.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotP, tt.wantP) {
				t.Errorf("New() = %v, want %v", gotP, tt.wantP)
			}
		})
	}
}

func TestProvider_CreateMachine(t *testing.T) {
	type fields struct {
		Host string
		Port int
	}
	tests := []struct {
		name    string
		fields  fields
		want    *iaas.Machine
		wantErr bool
	}{
		{
			"create machine",
			fields{
				"localhost",
				2376,
			},
			&iaas.Machine{
				IP:   "localhost",
				Port: 2376,
				Kind: "TCP",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				Host: tt.fields.Host,
				Port: tt.fields.Port,
			}
			got, err := p.CreateMachine()
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.CreateMachine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.CreateMachine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_DeleteMachine(t *testing.T) {
	type fields struct {
		Host string
		Port int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"delete machine",
			fields{
				"localhost",
				2376,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				Host: tt.fields.Host,
				Port: tt.fields.Port,
			}
			if err := p.DeleteMachine(); (err != nil) != tt.wantErr {
				t.Errorf("Provider.DeleteMachine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
