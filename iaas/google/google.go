package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"context"

	"github.com/docker/machine/drivers/google"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/gofn/gofn/iaas"
	uuid "github.com/satori/go.uuid"
)

// Provider definition, represents a concrete implementation of an iaas
type Provider struct {
	Client     libmachine.API
	Host       *host.Host
	Name       string
	ClientPath string
	Region     string
	Size       string
	ImageSlug  string
	KeyID      int
	Ctx        context.Context
}

type driverConfig struct {
	DriverName string `json:"DriverName"`
	Driver     struct {
		MachineName  string `json:"MachineName"`
		IPAddress    string `json:"IPAddress"`
		MachineImage string `json:"MachineImage"`
	} `json:"Driver"`
}

func getConfig(machineDir, hostName string) (config *driverConfig, err error) {
	configPath := fmt.Sprintf("%s/%s/config.json", machineDir, hostName)
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return
	}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return
	}
	return
}

func New(projectID string) (p *Provider, err error) {
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentials == "" {
		err = errors.New("You must set GOOGLE_APPLICATION_CREDENTIALS environment variable with the path for the credentials file")
		return
	}

	var uid uuid.UUID
	uid, err = uuid.NewV4()
	if err != nil {
		return
	}
	name := fmt.Sprintf("gofn-%s", uid.String())
	clientPath := "/tmp/" + name
	c := libmachine.NewClient(clientPath, clientPath+"/certs")
	driver := google.NewDriver(name, clientPath)
	driver.Project = projectID

	p = &Provider{
		Client:     c,
		Name:       name,
		ClientPath: clientPath,
	}

	data, err := json.Marshal(driver)
	if err != nil {
		return
	}

	p.Host, err = p.Client.NewHost(driver.DriverName(), data)
	if err != nil {
		return
	}
	return
}

// CreateMachine on google
func (p *Provider) CreateMachine() (machine *iaas.Machine, err error) {
	err = p.Client.Create(p.Host)
	if err != nil {
		return
	}
	config, err := getConfig(p.Client.GetMachinesDir(), p.Name)
	if err != nil {
		return
	}
	ip, err := p.Host.Driver.GetIP()
	if err != nil {
		return
	}

	machine = &iaas.Machine{
		ID:        "",
		IP:        ip,
		Image:     config.Driver.MachineImage,
		Kind:      config.DriverName,
		Name:      p.Name,
		SSHKeysID: []int{},
		CertsDir:  p.ClientPath + "/certs",
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (p *Provider) DeleteMachine() (err error) {
	err = p.Host.Driver.Remove()
	defer p.Client.Close()
	if err != nil {
		return
	}
	return
}
