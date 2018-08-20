package tcp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/gofn/gofn/iaas"
)

var clientPort int

type Provider struct {
	Client     libmachine.API
	Host       *host.Host
	Name       string
	ClientPath string
	ClientHost string
	ClientPort int
	Ctx        context.Context
}

// New create provider
func New(URL string) (p *Provider, err error) {
	u, err := url.Parse(URL)
	if err != nil {
		return
	}

	if u.Port() != "" {
		clientPort, err = strconv.Atoi(u.Port())
		if err != nil {
			return
		}
	}
	p = &Provider{
		ClientHost: u.Hostname(),
		ClientPort: clientPort,
	}
	return
}

// CreateMachine on remote Docker server
func (tcp *Provider) CreateMachine() (machine *iaas.Machine, err error) {
	machine = &iaas.Machine{
		IP:   tcp.ClientHost,
		Port: tcp.ClientPort,
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (tcp *Provider) DeleteMachine() (err error) {
	return
}
