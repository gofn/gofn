package iaas

// Iaas represents a infresture service
type Iaas interface {
	Auth() error
	CreateMachine() (*Machine, error)
	DeleteMachine(*Machine) error
	CreateSnapshot() error
	ExecCommand() ([]byte, error)
}

// Machine defines a generic machine
type Machine struct{}
