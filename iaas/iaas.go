package iaas

// Iaas represents a infresture service
type Iaas interface {
	Auth() error
	CreateMachine() (*Machine, error)
	DeleteMachine() error
	CreateSnapshot() error
	ExecCommand(cmd string) ([]byte, error)
}

// Machine defines a generic machine
type Machine struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	IP        string `json:"ip"`
	Image     string `json:"image"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	SSHKeysID []int  `json:"ssh_keys_id"`
}
