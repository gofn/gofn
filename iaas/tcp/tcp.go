package tcp

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofn/gofn/iaas"
)

// Provider definition, represents a concrete implementation of an iaas
type Provider struct {
	Host string
	Port int
}

var (
	errInvalidURL = errors.New("invalid TCP URL")
)

// New create provider
func New(URL string) (p *Provider, err error) {
	u, err := url.Parse(URL)
	if err != nil {
		return
	}
	if !strings.Contains(u.Scheme, "tcp") {
		err = errInvalidURL
		return
	}
	var clientPort int
	if u.Port() != "" {
		clientPort, err = strconv.Atoi(u.Port())
		if err != nil {
			return
		}
	}
	p = &Provider{
		Host: u.Hostname(),
		Port: clientPort,
	}
	return
}

// CreateMachine tcp iaas
func (p *Provider) CreateMachine() (*iaas.Machine, error) {
	return &iaas.Machine{
		IP:   p.Host,
		Port: p.Port,
		Kind: "TCP",
	}, nil
}

// DeleteMachine tcp iaas
func (p *Provider) DeleteMachine() error {
	return nil
}
