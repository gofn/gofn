package amazonec2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers/rpc"
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
		InstanceID  string `json:"InstanceID"`
		MachineName string `json:"MachineName"`
		IPAddress   string `json:"IPAddress"`
		AMI         string `json:"AMI"`
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

func setFlags(driver *amazonec2.Driver) (err error) {
	flags := driver.GetCreateFlags()

	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}

	// conversion of flags extracted from:
	// https://github.com/docker/machine/blob/master/commands/create.go#L354-L361
	for _, f := range flags {
		driverOpts.Values[f.String()] = f.Default()

		// Hardcoded logic for boolean... :(
		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}
	}
	// TODO: receive this configs to remove hard coded values
	driverOpts.Values["swarm-host"] = "tcp://0.0.0.0:3376"
	driverOpts.Values["swarm-image"] = "swarm:latest"
	driverOpts.Values["swarm-strategy"] = "spread"
	driverOpts.Values["swarm-discovery"] = ""
	driverOpts.Values["swarm-master"] = false
	err = driver.SetConfigFromFlags(&driverOpts)
	if err != nil {
		return
	}
	driver.SetSwarmConfigFromFlags(&driverOpts)
	return
}

func New(accessKey, secretKey string, opts ...iaas.ProviderOpts) (p *Provider, err error) {
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
	driver := amazonec2.NewDriver(p.Name, p.ClientPath)
	driver.AccessKey = accessKey
	driver.SecretKey = secretKey
	if p.ImageSlug != "" {
		driver.AMI = p.ImageSlug
	}
	if p.Region != "" {
		driver.Region = p.Region
	}
	if p.Size != "" {
		driver.InstanceType = p.Size
	}
	if p.KeyID != 0 {
		driver.SSHKeyID = p.KeyID
	}
	if p.DiskSize != 0 {
		driver.RootSize = int64(p.DiskSize)
	}
	err = setFlags(driver)
	if err != nil {
		p = nil
		return
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

// CreateMachine on digitalocean
func (p *Provider) CreateMachine() (machine *iaas.Machine, err error) {
	err = p.Client.Create(p.Host)
	if err != nil {
		return
	}
	config, err := getConfig(p.Client.GetMachinesDir(), p.Name)
	if err != nil {
		return
	}

	machine = &iaas.Machine{
		ID:        config.Driver.InstanceID,
		IP:        config.Driver.IPAddress,
		Image:     config.Driver.AMI,
		Kind:      config.DriverName,
		Name:      p.Name,
		SSHKeysID: []int{config.Driver.SSHKeyID},
		CertsDir:  p.ClientPath + "/certs",
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (p *Provider) DeleteMachine() (err error) {
	err = p.Host.Driver.Remove()
	defer p.Client.Close()
	return
}
