package iaas

const (
	// SmallRetry is the smallest observed value to retry the connection with the remote machine
	SmallRetry = 120
	// MediumRetry represents average value to retry the connection with the remote machine
	MediumRetry = 480
	// BigRetry is the biggest observed value to retry the connection with the remote machine
	BigRetry = 960
)

// Iaas represents a infresture service
type Iaas interface {
	CreateMachine() (*Machine, error)
	DeleteMachine(machine *Machine) error
	ExecCommand(machine *Machine, cmd string) ([]byte, error)
}

// Machine defines a generic machine
type Machine struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	Image     string `json:"image"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	SSHKeysID []int  `json:"ssh_keys_id"`
	CertsDir  string `json:certs_dir`
}
