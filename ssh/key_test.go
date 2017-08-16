package ssh

import (
	"io/ioutil"
	"testing"
)

func TestGeneratePrivateSSHKey(t *testing.T) {
	private, err := GeneratePrivateKey(32)
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
	if private == nil {
		t.Errorf("expected private not nil but is nil")
	}
}

func TestGeneratePublicSSHKey(t *testing.T) {
	private, err := GeneratePrivateKey(32)
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
	if private == nil {
		t.Errorf("expected private not nil but is nil")
	}
	err = GeneratePublicKey(private)
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
}

func TestGenerateFNSSHKey(t *testing.T) {
	err := GenerateFNSSHKey(32)
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
}

func TestGenerateFingerPrint(t *testing.T) {
	err := GenerateFNSSHKey(32)
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
	byt, err := ioutil.ReadFile(".gofn/keys/id_rsa.pub")
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
	f, err := GenerateFingerPrint(string(byt))
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
	if f == "" {
		t.Error("expected non empty string, but is empty")
	}
}
