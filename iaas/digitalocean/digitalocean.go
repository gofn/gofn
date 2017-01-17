package digitalocean

import (
	"errors"
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

func (do *Digitalocean) CreateMachine() (m *iaas.Machine, err error) {
	err = do.Auth()
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
	}
	newDroplet, resp, err := do.client.Droplets.Create(createRequest)
	if err != nil {
		return
	}
	ipv4, err := newDroplet.PublicIPv4()
	if err != nil {
		return
	}
	m = &iaas.Machine{
		ID:     strconv.Itoa(newDroplet.ID),
		IP:     ipv4,
		Image:  newDroplet.Image.Slug,
		Kind:   "digitalocean",
		Name:   newDroplet.Name,
		Status: newDroplet.Status,
	}
	return
}
