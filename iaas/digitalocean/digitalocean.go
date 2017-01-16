package digitalocean

import (
	"errors"
	"os"

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
	key := os.Getenv("DIGITALOCEAN_API_KEY")
	if key == "" {
		err = errors.New("You must provide a Digital Ocean API Key")
		return
	}
	tokenSource := &oauth2.StaticTokenSource{
		AccessToken: key,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	do.client = godo.NewClient(oauthClient)
	return
}
