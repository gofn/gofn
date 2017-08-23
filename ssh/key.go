package ssh

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

var (
	// KeysDir is the directory where keys are stored
	KeysDir = "./.gofn/keys"
	// PrivateKeyName is the default name of private key
	PrivateKeyName = "id_rsa"
	// PublicKeyName is the default name of public key
	PublicKeyName = "id_rsa.pub"
	// Port is the default ssh port
	Port = ":22"
)

// PublicKeyFile for auth method
func PublicKeyFile(file string) ssh.AuthMethod {
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

// WritePEM SSH
func WritePEM(path string, content []byte, filePermission os.FileMode, dirPermission os.FileMode) (err error) {
	err = os.MkdirAll(KeysDir, dirPermission)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path, content, filePermission)
	return
}

// GeneratePublicKey for SSH
func GeneratePublicKey(privateKey *rsa.PrivateKey) (err error) {
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	path := filepath.Join(KeysDir, PublicKeyName)
	err = WritePEM(path, ssh.MarshalAuthorizedKey(pub), 0644, 0700)
	return
}

// GeneratePrivateKey for SSH
func GeneratePrivateKey(bits int) (privateKey *rsa.PrivateKey, err error) {
	privateKey, _ = rsa.GenerateKey(rand.Reader, bits)
	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := pem.EncodeToMemory(&privateKeyBlock)
	path := filepath.Join(KeysDir, PrivateKeyName)
	err = WritePEM(path, privateKeyPem, 0600, 0700)
	return
}

// GenerateFNSSHKey func
func GenerateFNSSHKey(bits int) (err error) {
	privateKey, err := GeneratePrivateKey(bits)
	if err != nil {
		return
	}
	err = GeneratePublicKey(privateKey)
	return
}

// GenerateFingerPrint based content
func GenerateFingerPrint(content string) (fingerPrint string, err error) {
	parts := strings.Fields(content)
	if len(parts) < 2 {
		err = errors.New("bad content key")
		return
	}

	key, _ := base64.StdEncoding.DecodeString(parts[1])

	fp := md5.Sum(key)
	for i, b := range fp {
		fingerPrint += fmt.Sprintf("%02x", b)
		if i < len(fp)-1 {
			fingerPrint += ":"
		}
	}
	return
}
