package digitalocean

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/libmachine"
	"github.com/gofn/gofn/iaas"
	"github.com/gofrs/uuid"
)

// Provider definition, represents a concrete implementation of an iaas
type Provider struct {
	iaas.Provider
}

type driverConfig struct {
	Driver struct {
		DropletID   int    `json:"DropletID"`
		DropletName string `json:"DropletName"`
		IPAddress   string `json:"IPAddress"`
		Image       string `json:"Image"`
		SSHKeyID    int    `json:"SSHKeyID"`
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

func New(token string, opts ...iaas.ProviderOpts) (p *Provider, err error) {
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
	driver := digitalocean.NewDriver(name, clientPath)
	driver.AccessToken = token
	driver.Image = p.ImageSlug
	driver.Region = p.Region
	driver.Size = p.Size
	driver.SSHKeyID = p.KeyID
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

// CreateMachine on digitalocean
func (do *Provider) CreateMachine() (machine *iaas.Machine, err error) {
	err = do.Client.Create(do.Host)
	if err != nil {
		return
	}
	config, err := getConfig(do.Client.GetMachinesDir(), do.Name)
	if err != nil {
		return
	}

	machine = &iaas.Machine{
		ID:        strconv.Itoa(config.Driver.DropletID),
		IP:        config.Driver.IPAddress,
		Image:     config.Driver.Image,
		Kind:      "digitalocean",
		Name:      config.Driver.DropletName,
		SSHKeysID: []int{config.Driver.SSHKeyID},
		CertsDir:  do.ClientPath + "/certs",
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (do *Provider) DeleteMachine() (err error) {
	err = do.Host.Driver.Remove()
	defer do.Client.Close()
	if err != nil {
		return
	}
	return
}
