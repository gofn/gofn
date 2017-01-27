package digitalocean

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"crypto/rand"

	"github.com/digitalocean/godo"
	"github.com/nuveo/gofn/iaas"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

var (
	keysDir        = "./.gofn/keys"
	privateKeyName = "id_rsa"
	publicKeyName  = "id_rsa.pub"
	sshPort        = ":22"
)

// Digitalocean difinition
type Digitalocean struct {
	client *godo.Client
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
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
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
	snapshots, _, err := do.client.Snapshots.List(nil)
	if err != nil {
		return
	}

	snapshot := godo.Snapshot{}
	for _, s := range snapshots {
		if s.Name == "GOFN" {
			snapshot = s
			break
		}
	}
	image := godo.DropletCreateImage{
		Slug: "debian-8-x64",
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
		Region: "nyc1",
		Size:   "512mb",
		Image:  image,
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				ID:          sshKey.ID,
				Fingerprint: sshKey.Fingerprint,
			},
		},
	}
	newDroplet, _, err := do.client.Droplets.Create(createRequest)
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
		Status:    newDroplet.Status,
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

func writePEM(path string, content []byte, filePermission os.FileMode, dirPermission os.FileMode) (err error) {
	err = os.MkdirAll(keysDir, dirPermission)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path, content, filePermission)
	return
}

func generatePublicKey(privateKey *rsa.PrivateKey) (err error) {
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	path := filepath.Join(keysDir, publicKeyName)
	err = writePEM(path, ssh.MarshalAuthorizedKey(pub), 0644, 0700)
	return
}

func generatePrivateKey(bits int) (privateKey *rsa.PrivateKey, err error) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, bits)
	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := pem.EncodeToMemory(&privateKeyBlock)
	path := filepath.Join(keysDir, privateKeyName)
	err = writePEM(path, privateKeyPem, 0600, 0700)
	return
}

func generateFNSSHKey(bits int) (err error) {
	privateKey, err := generatePrivateKey(bits)
	if err != nil {
		return
	}
	err = generatePublicKey(privateKey)
	return
}

func (do *Digitalocean) getSSHKeyForDroplet() (sshKey *godo.Key, err error) {
	sshFilePath := os.Getenv("GOFN_SSH_PUBLICKEY_PATH")
	if sshFilePath == "" {
		path := filepath.Join(keysDir, publicKeyName)
		if !existsKey(path) {
			if err = generateFNSSHKey(4096); err != nil {
				return
			}
		}
		sshFilePath = path
	}
	content, err := ioutil.ReadFile(sshFilePath)
	if err != nil {
		return
	}
	strContent := strings.TrimSpace(string(content))
	sshKeys, _, err := do.client.Keys.List(nil)
	if err != nil {
		return
	}
	for _, key := range sshKeys {
		sshKey = &key
		if sshKey.PublicKey == strContent {
			return
		}
	}
	sshKeyRequestCreate := &godo.KeyCreateRequest{
		Name:      "GOFN",
		PublicKey: strContent,
	}
	sshKey, _, err = do.client.Keys.Create(sshKeyRequestCreate)
	if err != nil {
		return
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
	action, _, err := do.client.DropletActions.Shutdown(id)
	if err != nil {
		// Power off force Shutdown
		action, _, err = do.client.DropletActions.PowerOff(id)
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
			//rodando shutdown...
			select {
			case <-quit:
				return
			default:
				d, _, err := do.client.DropletActions.Get(id, action.ID)
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
		_, err = do.client.Droplets.Delete(id)
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
	action, _, err := do.client.DropletActions.Snapshot(id, "GOFN")
	if err != nil {
		return
	}
	timeout := 600
	quit := make(chan struct{})
	errs := make(chan error, 1)
	ac := make(chan *godo.Action, 1)
	go func() {
		for {
			//"rodando snapshot..."
			select {
			case <-quit:
				return
			default:
				d, _, err := do.client.DropletActions.Get(id, action.ID)
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

// ExecCommand on droplet
func (do *Digitalocean) ExecCommand(machine *iaas.Machine, cmd string) (output []byte, err error) {
	pkPath := os.Getenv("GOFN_SSH_PRIVATEKEY_PATH")
	if pkPath == "" {
		pkPath = filepath.Join(keysDir, privateKeyName)
	}
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			publicKeyFile(pkPath),
		},
		Timeout: time.Duration(5) * time.Minute,
	}
	conn, err := net.Dial("tcp", machine.IP+sshPort)
	for err != nil {
		conn, err = net.Dial("tcp", machine.IP+sshPort)
	}
	defer conn.Close()
	connection, err := ssh.Dial("tcp", machine.IP+sshPort, sshConfig)
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
				d, _, err := do.client.Droplets.Get(droplet.ID)
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
