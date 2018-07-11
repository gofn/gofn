package iaas

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
