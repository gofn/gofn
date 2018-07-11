package digitalocean

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"context"

	"github.com/docker/machine/drivers/digitalocean"
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

func New(token string) (do *Provider, err error) {
	var uid uuid.UUID
	uid, err = uuid.NewV4()
	if err != nil {
		return
	}
	name := fmt.Sprintf("gofn-%s", uid.String())
	clientPath := "/tmp/" + name
	c := libmachine.NewClient(clientPath, clientPath+"/certs")
	driver := digitalocean.NewDriver(name, clientPath)
	driver.AccessToken = token
	do = &Provider{
		Client:     c,
		Name:       name,
		ClientPath: clientPath,
	}

	data, err := json.Marshal(driver)
	if err != nil {
		return
	}

	do.Host, err = do.Client.NewHost(driver.DriverName(), data)
	if err != nil {
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
