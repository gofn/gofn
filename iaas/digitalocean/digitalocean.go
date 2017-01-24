package digitalocean

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"crypto/rand"

	"github.com/digitalocean/godo"
	"github.com/nuveo/gofn/iaas"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

var (
	keysDir        = "./.gofn/keys"
	privateKeyName = "id_rsa"
	publicKeyName  = "id_rsa.pub"
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
	for _, snapshot = range snapshots {
		if snapshot.Name == "GOFN" {
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
		Name:   "gofn",
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
	var cmd string
	if newDroplet.Image.Type != "snapshot" {

	}
	_, err = do.ExecCommand(machine, cmd)
	if err != nil {
		return
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
	publicKeyDer, _ := x509.MarshalPKIXPublicKey(&publicKey)

	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDer,
	}
	publicKeyPem := pem.EncodeToMemory(&publicKeyBlock)

	path := filepath.Join(keysDir, publicKeyName)
	err = writePEM(path, publicKeyPem, 0644, 0700)
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
	sshKeys, _, err := do.client.Keys.List(nil)
	if err != nil {
		return
	}
	for _, key := range sshKeys {
		sshKey = &key
		if sshKey.Name == "GOFN" {
			return
		}
	}
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
	sshKeyRequestCreate := &godo.KeyCreateRequest{
		Name:      "GOFN",
		PublicKey: string(content),
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
	_, _, err = do.client.DropletActions.Shutdown(id)
	if err != nil {
		// Power off force Shutdown
		_, _, err = do.client.DropletActions.PowerOff(id)
		if err != nil {
			return
		}
	}
	_, err = do.client.Droplets.Delete(id)
	return
}

// CreateSnapshot Create a snapshot from the machine
func (do *Digitalocean) CreateSnapshot(machine *iaas.Machine) (err error) {
	id, _ := strconv.Atoi(machine.ID)
	err = do.Auth()
	if err != nil {
		return
	}
	_, _, err = do.client.DropletActions.Snapshot(id, "GOFN")
	if err != nil {
		return
	}
	return
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
	pkPath := os.Getenv("GO_FN_PRIVATEKEY_PATH")
	if pkPath == "" {
		pkPath = filepath.Join(keysDir, privateKeyName)
	}
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			publicKeyFile(pkPath),
		},
	}
	connection, err := ssh.Dial("tcp", machine.IP, sshConfig)
	if err != nil {
		return
	}
	session, err := connection.NewSession()
	if err != nil {
		return
	}
	output, err = session.CombinedOutput(cmd)
	if err != nil {
		return
	}
	return
}

func existsKey(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
