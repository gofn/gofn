package digitalocean

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"context"

	"github.com/digitalocean/godo"
	"github.com/nuveo/gofn/iaas"
	gofnssh "github.com/nuveo/gofn/ssh"
	"github.com/nuveo/log"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

const (
	defaultRegion       = "nyc3"
	defaultSize         = "512mb"
	defaultImageSlug    = "debian-8-x64"
	defaultSnapshotName = "GOFN"
)

var (
	// ErrSnapshotNotFound is the error returned if
	// ErrorIfSnapshotNotExist is true and there is no snapshot
	// with name defined in SnapshotName
	ErrSnapshotNotFound = errors.New("snapshot not found")
)

// Digitalocean definition, represents a concrete implementation of an iaas
type Digitalocean struct {
	client    *godo.Client
	Region    string
	Size      string
	ImageSlug string
	KeyID     int
	Ctx       context.Context
	// SnapshotName if not defined GOFN will be used.
	SnapshotName string
	// ErrorIfSnapshotNotExist if true CreateMachine
	// returns error if a snapshot does not exist,
	// if false the system will try to create a snapshot,
	// defalt false.
	ErrorIfSnapshotNotExist bool
	sshPublicKeyPath        string
	sshPrivateKeyPath       string
}

// SetSSHPublicKeyPath adjust the system path for the ssh key
// but if the environment variable GOFN_SSH_PUBLICKEY_PATH exists
// the system will use the value contained in the variable instead
// of the one entered in SetSSHPublicKeyPath
func (do *Digitalocean) SetSSHPublicKeyPath(path string) {
	do.sshPublicKeyPath = path
}

// SetSSHPrivateKeyPath adjust the system path for the ssh key
// but if the environment variable GOFM_SSH_PRIVATEKEY_PATH exists
// the system will use the value contained in the variable instead
// of the one entered in SetSSHPrivateKeyPath
func (do *Digitalocean) SetSSHPrivateKeyPath(path string) {
	do.sshPrivateKeyPath = path
}

// GetSSHPublicKeyPath the path may change according to the
// environment variable GOFN_SSH_PUBLICKEY_PATH or by using
// the SetSSHPublicKeyPath
func (do *Digitalocean) GetSSHPublicKeyPath() (path string) {
	path = os.Getenv("GOFN_SSH_PUBLICKEY_PATH")
	if path != "" {
		return
	}
	path = do.sshPublicKeyPath
	if path != "" {
		return
	}
	do.sshPublicKeyPath = filepath.Join(gofnssh.KeysDir, gofnssh.PublicKeyName)
	path = do.sshPublicKeyPath
	return
}

// GetSSHPrivateKeyPath the path may change according to the
// environment variable GOFM_SSH_PRIVATEKEY_PATH or by using
// the SetSSHPrivateKeyPath
func (do *Digitalocean) GetSSHPrivateKeyPath() (path string) {
	path = os.Getenv("GOFN_SSH_PRIVATEKEY_PATH")
	if path != "" {
		return
	}
	path = do.sshPrivateKeyPath
	if path != "" {
		return
	}
	do.sshPrivateKeyPath = filepath.Join(gofnssh.KeysDir, gofnssh.PrivateKeyName)
	path = do.sshPrivateKeyPath
	return
}

// GetSnapshotName returns snapshot name or default if empty
func (do Digitalocean) GetSnapshotName() string {
	if do.SnapshotName == "" {
		return defaultSnapshotName
	}
	return do.SnapshotName
}

// GetRegion returns region or default if empty
func (do Digitalocean) GetRegion() string {
	if do.Region == "" {
		return defaultRegion
	}
	return do.Region
}

// GetSize returns size or default if empty
func (do Digitalocean) GetSize() string {
	if do.Size == "" {
		return defaultSize
	}
	return do.Size
}

// GetImageSlug returns image slug  or default if empty
func (do Digitalocean) GetImageSlug() string {
	if do.ImageSlug == "" {
		return defaultImageSlug
	}
	return do.ImageSlug
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
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	do.client = godo.NewClient(oauthClient)
	if apiURL != "" {
		do.client.BaseURL, err = url.Parse(apiURL)
		if err != nil {
			return
		}
	}
	return
}

// CreateMachine on digitalocean
func (do *Digitalocean) CreateMachine() (machine *iaas.Machine, err error) {
	err = do.Auth()
	if err != nil {
		return
	}
	lo := godo.ListOptions{
		Page:    1,
		PerPage: 999999,
	}
	snapshots, _, err := do.client.Snapshots.List(do.Ctx, &lo)
	if err != nil {
		return
	}
	snapshot := godo.Snapshot{}
	for _, s := range snapshots {
		if s.Name == do.GetSnapshotName() {
			snapshot = s
			break
		}
	}
	if snapshot.Name == "" && do.ErrorIfSnapshotNotExist {
		err = ErrSnapshotNotFound
		return
	}
	image := godo.DropletCreateImage{
		Slug: do.GetImageSlug(),
	}
	if snapshot.Name != "" {
		id, _ := strconv.Atoi(snapshot.ID)
		image = godo.DropletCreateImage{
			ID: id,
		}
	}
	sshKey, err := do.getSSHKeyForDroplet()
	if err != nil {
		return
	}
	createRequest := &godo.DropletCreateRequest{
		Name:   fmt.Sprintf("gofn-%s", uuid.NewV4().String()),
		Region: do.GetRegion(),
		Size:   do.GetSize(),
		Image:  image,
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				ID:          sshKey.ID,
				Fingerprint: sshKey.Fingerprint,
			},
		},
	}
	newDroplet, _, err := do.client.Droplets.Create(do.Ctx, createRequest)
	if err != nil {
		return
	}
	newDroplet, err = do.waitNetworkCreated(newDroplet)
	if err != nil {
		return
	}
	ipv4, err := newDroplet.PublicIPv4()
	if err != nil {
		return
	}
	machine = &iaas.Machine{
		ID:        strconv.Itoa(newDroplet.ID),
		IP:        ipv4,
		Image:     newDroplet.Image.Slug,
		Kind:      "digitalocean",
		Name:      newDroplet.Name,
		SSHKeysID: []int{sshKey.ID},
	}
	cmd := fmt.Sprintf(iaas.RequiredDeps, machine.IP)
	if snapshot.Name == "" {
		cmd = iaas.OptionalDeps + cmd
	}
	_, err = do.ExecCommand(machine, cmd)
	if err != nil {
		return
	}
	if snapshot.Name == "" {
		err = do.CreateSnapshot(machine)
		if err != nil {
			return
		}
	}
	return
}

func (do *Digitalocean) getSSHKeyForDroplet() (sshKey *godo.Key, err error) {
	// Use a key that is already in DO if exist KeyID
	if do.KeyID != 0 {
		sshKey, _, err = do.client.Keys.GetByID(do.Ctx, do.KeyID)
		if err != nil {
			return
		}
		return
	}
	sshFilePath := do.GetSSHPublicKeyPath()
	if sshFilePath == "" {
		path := filepath.Join(gofnssh.KeysDir, gofnssh.PublicKeyName)
		if !existsKey(path) {
			if err = gofnssh.GenerateFNSSHKey(4096); err != nil {
				return
			}
		}
		sshFilePath = path
	}
	content, err := ioutil.ReadFile(sshFilePath)
	if err != nil {
		return
	}

	fingerPrint, err := gofnssh.GenerateFingerPrint(string(content))
	if err != nil {
		return
	}

	sshKey, _, err = do.client.Keys.GetByFingerprint(do.Ctx, fingerPrint)
	if err != nil {
		sshKeyRequestCreate := &godo.KeyCreateRequest{
			Name:      "GOFN",
			PublicKey: string(content),
		}
		sshKey, _, err = do.client.Keys.Create(do.Ctx, sshKeyRequestCreate)
		if err != nil {
			return
		}
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (do *Digitalocean) DeleteMachine(machine *iaas.Machine) (err error) {
	id, _ := strconv.Atoi(machine.ID)
	err = do.Auth()
	if err != nil {
		return
	}
	action, _, err := do.client.DropletActions.Shutdown(do.Ctx, id)
	if err != nil {
		log.Println(err)
		// Power off force Shutdown
		action, _, err = do.client.DropletActions.PowerOff(do.Ctx, id)
		if err != nil {
			return
		}
	}
	timeout := 120
	quit := make(chan struct{})
	errs := make(chan error, 1)
	ac := make(chan *godo.Action, 1)
	go func() {
		for {
			//running shutdown...
			select {
			case <-quit:
				return
			default:
				var d *godo.Action
				d, _, err = do.client.DropletActions.Get(do.Ctx, id, action.ID)
				if err != nil {
					errs <- err
					return
				}
				if d.Status == "completed" {
					ac <- d
					return
				}
			}
		}
	}()
	select {
	case action = <-ac:
		_, err = do.client.Droplets.Delete(do.Ctx, id)
		return
	case err = <-errs:
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		err = errors.New("timed out waiting for Snhutdown")
		return
	}
}

// CreateSnapshot Create a snapshot from the machine
func (do *Digitalocean) CreateSnapshot(machine *iaas.Machine) (err error) {
	id, _ := strconv.Atoi(machine.ID)
	err = do.Auth()
	if err != nil {
		return
	}
	action, _, err := do.client.DropletActions.Snapshot(do.Ctx, id, do.GetSnapshotName())
	if err != nil {
		return
	}
	timeout := 600
	quit := make(chan struct{})
	errs := make(chan error, 1)
	ac := make(chan *godo.Action, 1)
	go func() {
		for {
			//"running snapshot..."
			select {
			case <-quit:
				return
			default:
				var d *godo.Action
				d, _, err = do.client.DropletActions.Get(do.Ctx, id, action.ID)
				if err != nil {
					errs <- err
					return
				}
				if d.Status == "completed" {
					ac <- d
					return
				}
			}
		}
	}()
	select {
	case action = <-ac:
		return
	case err = <-errs:
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		err = errors.New("timed out waiting for Snapshot")
		return
	}
}

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func probeConnection(ip string, maxRetries int) error {
	counter := 0
	var (
		conn net.Conn
		err  error
	)
	for counter < maxRetries {
		conn, err = net.DialTimeout("tcp", ip+gofnssh.Port, time.Duration(500)*time.Millisecond)
		if err == nil {
			return nil
		}
		counter++
		time.Sleep(time.Duration(250) * time.Millisecond)
	}

	if conn != nil {
		err = conn.Close()
	}
	return err
}

// ExecCommand on droplet
func (do *Digitalocean) ExecCommand(machine *iaas.Machine, cmd string) (output []byte, err error) {
	pkPath := do.GetSSHPrivateKeyPath()

	// TODO: dynamic user
	sshConfig := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			publicKeyFile(pkPath),
		},
		Timeout: time.Duration(10) * time.Second,
	}

	err = probeConnection(machine.IP, iaas.MediumRetry)
	if err != nil {
		return
	}
	connection, err := ssh.Dial("tcp", machine.IP+gofnssh.Port, sshConfig)
	if err != nil {
		return
	}
	session, err := connection.NewSession()
	if err != nil {
		return
	}
	output, err = session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(string(output))
		return
	}
	return
}

func (do *Digitalocean) waitNetworkCreated(droplet *godo.Droplet) (upDroplet *godo.Droplet, err error) {
	timeout := 120
	quit := make(chan struct{})
	errs := make(chan error, 1)
	droplets := make(chan *godo.Droplet, 1)
	go func() {
		for {
			//wait for network
			select {
			case <-quit:
				return
			default:
				d, _, err := do.client.Droplets.Get(do.Ctx, droplet.ID)
				if err != nil {
					errs <- err
					return
				}
				if len(d.Networks.V4) > 0 && !d.Locked {
					droplets <- d
					return
				}
			}
		}
	}()
	select {
	case upDroplet = <-droplets:
		return upDroplet, nil
	case err := <-errs:
		return nil, err
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, errors.New("timed out waiting for machine network")
	}
}

func existsKey(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
