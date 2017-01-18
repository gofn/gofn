package digitalocean

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"

	"github.com/digitalocean/godo"
	"github.com/nuveo/gofn/iaas"
	"golang.org/x/oauth2"
)

// Digitalocean difinition
type Digitalocean struct {
	iaas.Machine
	client *godo.Client
}

// Auth in Digitalocean API
func (do *Digitalocean) Auth() (err error) {
	apiURL := os.Getenv("DIGITALOCEAN_API_URL")
	key := os.Getenv("DIGITALOCEAN_API_KEY")
	if key == "" {
		err = errors.New("You must provide a Digital Ocean API Key")
		return
	}
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: key,
	})
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	do.client = godo.NewClient(oauthClient)
	if apiURL != "" {
		do.client.BaseURL, err = url.Parse(apiURL)
		if err != nil {
			return
		}
	}
	return
}

// CreateMachine on digitalocean
func (do *Digitalocean) CreateMachine() (m *iaas.Machine, err error) {
	err = do.Auth()
	if err != nil {
		return
	}
	sshKey, err := do.getSSHKeyForDroplet()
	if err != nil {
		return
	}
	createRequest := &godo.DropletCreateRequest{
		Name:   "gofn",
		Region: "nyc3",
		Size:   "512mb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-16-10-x64",
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				ID:          sshKey.ID,
				Fingerprint: sshKey.Fingerprint,
			},
		},
	}
	newDroplet, _, err := do.client.Droplets.Create(createRequest)
	if err != nil {
		return
	}
	ipv4, err := newDroplet.PublicIPv4()
	if err != nil {
		return
	}
	m = &iaas.Machine{
		ID:        strconv.Itoa(newDroplet.ID),
		IP:        ipv4,
		Image:     newDroplet.Image.Slug,
		Kind:      "digitalocean",
		Name:      newDroplet.Name,
		Status:    newDroplet.Status,
		SSHKeysID: []int{sshKey.ID},
	}
	return
}

func (do *Digitalocean) getSSHKeyForDroplet() (sshKey *godo.Key, err error) {
	sshKeys, _, err := do.client.Keys.List(nil)
	if err != nil {
		return
	}
	for _, key := range sshKeys {
		sshKey = &key
		if sshKey.Name == "Gofn" {
			return
		}
	}
	sshFilePath := os.Getenv("GOFN_SSH_FILE_PATH")
	if sshFilePath == "" {
		err = errors.New("You must provide a SSH file path")
		return
	}
	content, err := ioutil.ReadFile(sshFilePath)
	if err != nil {
		return
	}
	sshKeyRequestCreate := &godo.KeyCreateRequest{
		Name:      "Gofn",
		PublicKey: string(content),
	}
	sshKey, _, err = do.client.Keys.Create(sshKeyRequestCreate)
	if err != nil {
		return
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (do *Digitalocean) DeleteMachine(mac *iaas.Machine) (err error) {
	id, _ := strconv.Atoi(mac.ID)
	err = do.Auth()
	if err != nil {
		return
	}
	_, _, err = do.client.DropletActions.Shutdown(id)
	if err != nil {
		// Power off force Shutdown
		_, _, err = do.client.DropletActions.PowerOff(id)
		if err != nil {
			return
		}
	}
	_, err = do.client.Droplets.Delete(id)
	return
}
