package iaas

const (
	// RequiredDeps are depencies required to expose docker tcp service
	RequiredDeps = `mkdir -p  /etc/systemd/system/docker.service.d/
echo """
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://%s:2375 -H unix:///var/run/docker.sock
""" > /etc/systemd/system/docker.service.d/custom.conf
systemctl daemon-reload
systemctl restart docker
`
	// OptionalDeps are depencies required once, needed to setup docker services in a raw machine
	OptionalDeps = `curl https://raw.githubusercontent.com/nuveo/boxos/master/initial.sh | sh
sed -i  's/fd:\/\//fd:\/\/ $DOCKER_OPTS/g' /lib/systemd/system/docker.service
`
	// SmallRetry is the smallest observed value to retry the connection with the remote machine
	SmallRetry = 120
	// MediumRetry represents average value to retry the connection with the remote machine
	MediumRetry = 480
	// BigRetry is the biggest observed value to retry the connection with the remote machine
	BigRetry = 960
)

// Iaas represents a infresture service
type Iaas interface {
	Auth() error
	CreateMachine() (*Machine, error)
	DeleteMachine(machine *Machine) error
	CreateSnapshot(machine *Machine) error
	SetSSHPublicKeyPath(string)
	SetSSHPrivateKeyPath(string)
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
}
