package amazonec2

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/gofn/gofn/iaas"
)

func Test_getConfig(t *testing.T) {
	type args struct {
		machineDir string
		hostName   string
	}
	tests := []struct {
		name       string
		args       args
		wantConfig *driverConfig
		wantErr    bool
	}{
		{name: "config not found", args: args{machineDir: "./testdata/", hostName: "notfound"}, wantErr: true},
		{name: "problem to parse json", args: args{machineDir: "./testdata/", hostName: "unparseable"}, wantErr: true},
		{name: "correct parser", args: args{machineDir: "./testdata/", hostName: "testconfig"}, wantConfig: &driverConfig{
			DriverName: "amazonec2",
			Driver: struct {
				InstanceID  string "json:\"InstanceID\""
				MachineName string "json:\"MachineName\""
				IPAddress   string "json:\"IPAddress\""
				AMI         string "json:\"AMI\""
				SSHKeyID    int    "json:\"SSHKeyID\""
			}{
				InstanceID:  "i-00000000000000",
				MachineName: "gofn-test",
				IPAddress:   "111.222.333.444",
				AMI:         "ami-927185ef",
				SSHKeyID:    999999,
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := getConfig(tt.args.machineDir, tt.args.hostName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotConfig, tt.wantConfig) {
				t.Errorf("getConfig() = %#v, want %v", gotConfig, tt.wantConfig)
			}
		})
	}
}

type faultyDriver struct {
	fakedriver.Driver
}

func (m *faultyDriver) PreCreateCheck() error {
	return errors.New("Error in pre check mock")
}

func (m *faultyDriver) DriverName() string {
	return "generic"
}

type myAPI struct {
	libmachinetest.FakeAPI
}

func (m *myAPI) GetMachinesDir() string {
	return "./testdata"
}

func TestCreateMachine(t *testing.T) {
	// error on create machine
	p := Provider{
		iaas.Provider{
			Client: libmachine.NewClient("", ""),
		},
	}
	driver := &faultyDriver{}
	data, err := json.Marshal(driver)
	if err != nil {
		t.Fatal(err)
	}
	p.Host, err = p.Client.NewHost(driver.DriverName(), data)
	if err != nil {
		t.Fatal(err)
	}
	p.Host.Driver = driver
	_, err = p.CreateMachine()
	if err == nil {
		t.Fatal(err)
	}
	// error on get config
	p = Provider{
		iaas.Provider{
			Client: &libmachinetest.FakeAPI{},
		},
	}
	driver2 := &fakedriver.Driver{}
	p.Host = &host.Host{}
	p.Host.Driver = driver2
	_, err = p.CreateMachine()
	if err == nil {
		t.Fatal(err)
	}
	// sucess test
	p = Provider{
		iaas.Provider{
			Client: &myAPI{},
		},
	}
	p.Name = "testconfig"
	driver3 := &fakedriver.Driver{}
	p.Host = &host.Host{}
	p.Host.Driver = driver3
	_, err = p.CreateMachine()
	if err != nil {
		t.Fatal(err)
	}

}

type deleteAPI struct {
	libmachinetest.FakeAPI
}

func (d deleteAPI) Close() error {
	return errors.New("This error will be ignored")
}

type removeDriver struct {
	fakedriver.Driver
}

func (r removeDriver) Remove() error {
	return errors.New("error on remove")
}

func TestDeleteMachine(t *testing.T) {
	// success
	p := Provider{
		iaas.Provider{
			Client: &libmachinetest.FakeAPI{},
		},
	}
	driver := &fakedriver.Driver{}
	p.Host = &host.Host{}
	p.Host.Driver = driver
	err := p.DeleteMachine()
	if err != nil {
		t.Fatal(err)
	}
	// error on close will be ignored
	p = Provider{
		iaas.Provider{
			Client: &deleteAPI{},
		},
	}
	p.Host = &host.Host{}
	p.Host.Driver = driver
	err = p.DeleteMachine()
	if err != nil {
		t.Fatal(err)
	}
	// error on remove
	p = Provider{
		iaas.Provider{
			Client: &libmachinetest.FakeAPI{},
		},
	}
	driver2 := &removeDriver{}
	p.Host = &host.Host{}
	p.Host.Driver = driver2
	err = p.DeleteMachine()
	if err == nil {
		t.Fatal(err)
	}
}
