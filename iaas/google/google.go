package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/machine/drivers/google"
	"github.com/docker/machine/libmachine"
	"github.com/gofn/gofn/iaas"
	"github.com/gofrs/uuid"
)

// Provider definition, represents a concrete implementation of an iaas
type Provider struct {
	iaas.Provider
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

func New(projectID string, opts ...iaas.ProviderOpts) (p *Provider, err error) {
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentials == "" {
		err = errors.New("you must set GOOGLE_APPLICATION_CREDENTIALS environment variable with the path for the credentials file")
		return
	}
	p = &Provider{}
	for _, opt := range opts {
		if err = opt(&p.Provider); err != nil {
			p = nil
			return
		}
	}
	var uid uuid.UUID
	uid, err = uuid.NewV4()
	if err != nil {
		p = nil
		return
	}
	name := fmt.Sprintf("gofn-%s", uid.String())
	if p.Name == "" {
		p.Name = name
	}
	clientPath := "/tmp/" + p.Name
	if p.ClientPath == "" {
		p.ClientPath = clientPath
	}
	p.Client = libmachine.NewClient(p.ClientPath, p.ClientPath+"/certs")
	driver := google.NewDriver(name, clientPath)
	driver.Project = projectID
	p.ImageSlug = strings.TrimPrefix(p.ImageSlug, "https://www.googleapis.com/compute/v1/projects/")
	if p.ImageSlug != "" {
		driver.MachineImage = p.ImageSlug
	}
	if p.Region != "" {
		driver.Zone = p.Region
	}
	if p.Size != "" {
		driver.MachineType = p.Size
	}
	if p.DiskSize != 0 {
		driver.DiskSize = p.DiskSize
	}
	data, err := json.Marshal(driver)
	if err != nil {
		p = nil
		return
	}
	p.Host, err = p.Client.NewHost(driver.DriverName(), data)
	if err != nil {
		p = nil
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
