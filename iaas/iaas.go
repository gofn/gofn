package iaas

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
)

// Iaas represents a infresture service
type Iaas interface {
	CreateMachine() (*Machine, error)
	DeleteMachine() error
}

// Machine defines a generic machine
type Machine struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	Image     string `json:"image"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	SSHKeysID []int  `json:"ssh_keys_id"`
	CertsDir  string `json:"certs_dir"`
}

// Provider for gofn
type Provider struct {
	Client     libmachine.API
	Host       *host.Host
	Name       string
	ClientPath string
	Region     string
	Size       string
	ImageSlug  string
	KeyID      int
	DiskSize   int
}

// ProviderOpts override defaults
type ProviderOpts func(*Provider) error

// WithName func
func WithName(name string) ProviderOpts {
	return func(p *Provider) error {
		p.Name = name
		return nil
	}
}

// WithSO func
func WithSO(so string) ProviderOpts {
	return func(p *Provider) error {
		p.ImageSlug = so
		return nil
	}
}

// WithSize func
func WithSize(size string) ProviderOpts {
	return func(p *Provider) error {
		p.Size = size
		return nil
	}
}

// WithDiskSize func
func WithDiskSize(size int) ProviderOpts {
	return func(p *Provider) error {
		p.DiskSize = size
		return nil
	}
}

// WithKeyID func
func WithKeyID(keyID int) ProviderOpts {
	return func(p *Provider) error {
		p.KeyID = keyID
		return nil
	}
}

// WithRegion func
func WithRegion(region string) ProviderOpts {
	return func(p *Provider) error {
		p.Region = region
		return nil
	}
}

// WithClientPath func
func WithClientPath(path string) ProviderOpts {
	return func(p *Provider) error {
		p.ClientPath = path
		return nil
	}
}
