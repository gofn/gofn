package digitalocean

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"crypto/rand"

	"github.com/digitalocean/godo"
	"github.com/nuveo/gofn/iaas"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

const (
	keysDir = "./.gofn/keys"
	keyName = "id_rsa"
)

// Digitalocean difinition
type Digitalocean struct {
	iaas.Machine
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
func (do *Digitalocean) CreateMachine() (m *iaas.Machine, err error) {
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
		if snapshot.Name == "Gofn" {
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
	m = &iaas.Machine{
		ID:        strconv.Itoa(newDroplet.ID),
		IP:        ipv4,
		Image:     newDroplet.Image.Slug,
		Kind:      "digitalocean",
		Name:      newDroplet.Name,
		Status:    newDroplet.Status,
		SSHKeysID: []int{sshKey.ID},
	}
	return
}

func generateFNSSHKey() (err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}
	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := pem.EncodeToMemory(&privateKeyBlock)
	path := filepath.Join(keysDir, keyName)
	err = os.MkdirAll(keysDir, 0644)
	if err != nil {
		return
	}
	ioutil.WriteFile(path, privateKeyPem, 0644)
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
		path := filepath.Join(keysDir, keyName)
		if !existsKey(path) {
			if err = generateFNSSHKey(); err != nil {
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
		Name:      "Gofn",
		PublicKey: string(content),
	}
	sshKey, _, err = do.client.Keys.Create(sshKeyRequestCreate)
	if err != nil {
		return
	}
	return
}

// DeleteMachine Shutdown and Delete a droplet
func (do *Digitalocean) DeleteMachine(mac *iaas.Machine) (err error) {
	id, _ := strconv.Atoi(mac.ID)
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
func (do *Digitalocean) CreateSnapshot(mac *iaas.Machine) (err error) {
	id, _ := strconv.Atoi(mac.ID)
	err = do.Auth()
	if err != nil {
		return
	}
	_, _, err = do.client.DropletActions.Snapshot(id, "Gofn")
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

func (do *Digitalocean) ExecCommand(cmd string) (output []byte, err error) {
	pkPath := os.Getenv("GO_FN_PRIVATEKEY_PATH")
	if pkPath == "" {
		var usr *user.User
		usr, err = user.Current()
		if err != nil {
			return
		}
		pkPath = filepath.Join(usr.HomeDir, "/.gofn/keys/id_rsa")
	}
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			publicKeyFile(pkPath),
		},
	}
	connection, err := ssh.Dial("tcp", do.IP, sshConfig)
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
