package ssh

import (
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestFakeServer(t *testing.T) {
	server, err := NewServer(t, "127.0.0.1:4400", "echo test", "test", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	pkPath := filepath.Join(KeysDir, PrivateKeyName)
	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			PublicKeyFile(pkPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", server.Address(), sshConfig)
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	output, err := session.CombinedOutput(server.Cmd)
	if err != nil {
		t.Error(err)
	}
	if string(output) != server.Reply {
		t.Errorf("expected %v, but got %v", server.Reply, string(output))
	}
}
