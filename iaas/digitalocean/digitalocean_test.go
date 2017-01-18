package digitalocean

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nuveo/gofn/iaas"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	os.Setenv("DIGITALOCEAN_API_URL", server.URL)
	os.Setenv("DIGITALOCEAN_API_KEY", "api-key")
}

func teardown() {
	server.Close()
}

func defineListSnapshotsEndpoint() {
	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		snap := `{"snapshots": [
								{
								  "id": "6372321",
								  "name": "Gofn",
								  "regions": [
									"nyc1",
									"ams1",
									"sfo1",
									"nyc2",
									"ams2",
									"sgp1",
									"lon1",
									"nyc3",
									"ams3",
									"fra1",
									"tor1"
								  ],
								  "created_at": "2014-09-26T16:40:18Z",
								  "resource_id": "2713828",
								  "resource_type": "droplet",
								  "min_disk_size": 20,
								  "size_gigabytes": 1.42
								}]
					}`
		fmt.Fprint(w, snap)
	})
}

func defineBrokenListSnapshotsEndpoint() {
	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		snap := `{"snapshots": [
								{
								  "id": "6372321",
								  "name": "5.10 x64",
								  "regions": [
									"nyc1",
									"ams1",
									"sfo1",
									"nyc2",
									"ams2",
									"sgp1",
									"lon1",
									"nyc3",
									"ams3",
									"fra1",
									"tor1"
								  ],
								  "created_at": "2014-09-26T16:40:18Z",
								  "resource_id": "2713828",
								  "resource_type": "droplet",
								  "min_disk_size": 20,
								  "size_gigabytes": 1.42,
								}]
					}`
		fmt.Fprint(w, snap)
	})
}

func TestAuth(t *testing.T) {
	for _, test := range []struct {
		apiKEY   string
		apiURL   string
		baseURL  string
		errIsNil bool
	}{
		{"", "", "", false},
		{"apikey", "", "https://api.digitalocean.com/", true},
		{"apikey", "http://127.0.0.1:3000", "http://127.0.0.1:3000", true},
		{"apikey", "://localhost", "", false},
	} {
		do := &Digitalocean{}
		os.Setenv("DIGITALOCEAN_API_KEY", test.apiKEY)
		os.Setenv("DIGITALOCEAN_API_URL", test.apiURL)
		errBool := do.Auth() == nil
		if errBool != test.errIsNil {
			t.Errorf("%+v Expected %+v but found %+v", test, test.errIsNil, errBool)
		}
		if errBool && (do.client.BaseURL.String() != test.baseURL) {
			t.Errorf("Expected %q but found %q", test.baseURL, do.client.BaseURL)
		}
	}
}

func TestCreateMachine(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "testdata/fake_ssh.pub")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "Gofn"
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "Gofn"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	m, err := do.CreateMachine()
	if err != nil {
		t.Fatalf("Expected run without errors but has %q", err)
	}
	if m.ID != "1" {
		t.Errorf("Expected id = 1 but found %s", m.ID)
	}
	if m.IP != "104.131.186.241" {
		t.Errorf("Expected id = 104.131.186.241 but found %s", m.IP)
	}
	if m.Name != "gofn" {
		t.Errorf("Expected name = \"gofn\" but found %q", m.Name)
	}
	if m.Status != "new" {
		t.Errorf("Expected status = \"new\" but found %q", m.Status)
	}
	if m.SSHKeysID[0] != 512189 {
		t.Errorf("Expected SSHKeysID = 512189 but found %q", m.SSHKeysID[0])
	}
}

func TestCreateMachineWrongAuth(t *testing.T) {
	os.Setenv("DIGITALOCEAN_API_URL", "://localhost")
	do := &Digitalocean{}
	m, err := do.CreateMachine()
	if err == nil || m != nil {
		t.Errorf("expected erros but run without errors")
	}
}

func TestCreateMachineWrongIP(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "Gofn"
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "Gofn"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Errorf("expected errors but run without errors")
	}
}

func TestCreateMachineRequestError(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "Gofn"
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "Gofn"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Errorf("expected errors but run without errors")
	}
}

func TestCreateMachineWithNewSSHKey(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "testdata/fake_ssh.pub")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "my_key"
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)

			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "my_key"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	m, err := do.CreateMachine()
	if err != nil {
		t.Fatalf("Expected run without errors but has %q", err)
	}
	if m.ID != "1" {
		t.Errorf("Expected id = 1 but found %s", m.ID)
	}
	if m.IP != "104.131.186.241" {
		t.Errorf("Expected id = 104.131.186.241 but found %s", m.IP)
	}
	if m.Name != "gofn" {
		t.Errorf("Expected name = \"gofn\" but found %q", m.Name)
	}
	if m.Status != "new" {
		t.Errorf("Expected status = \"new\" but found %q", m.Status)
	}
	if m.SSHKeysID[0] != 512189 {
		t.Errorf("Expected SSHKeysID = 512189 but found %q", m.SSHKeysID[0])
	}
}

func TestCreateMachineWithWrongSSHKey(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "testdata/fake_ssh.pub")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "my_key",
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)

			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "my_key"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Fatalf("Expected run with errors but not has %q", err)
	}
}

func TestCreateMachineWithWrongSSHKeyList(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "testdata/fake_ssh.pub")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "my_key"
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)

			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "my_key",
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Fatalf("Expected run with errors but not has %q", err)
	}
}

func TestCreateMachineWithoutSSHKey(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "my_key",
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)

			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "my_key"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Fatalf("Expected run with errors but not has %q", err)
	}
}

func TestCreateMachineWithWrongSSHKeyPath(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "test/bla.pub")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "my_key",
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)

			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "my_key"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Fatalf("Expected run with errors but not has %q", err)
	}
}

func TestCreateMachineWrongSnapshotList(t *testing.T) {
	setup()
	defer teardown()
	os.Setenv("GOFN_SSH_FILE_PATH", "testdata/fake_ssh.pub")
	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method POST but request method is %s", r.Method)
		}
		droplet := `{"droplet": {
						"id": 1,
						"name": "gofn",
						"region": {"slug": "nyc3"},
						"status": "new",
						"image": {"slug": "ubuntu-16-10-x64"},
						"networks": {
							"v4":[
								{
									"ip_address": "104.131.186.241",
									"type": "public"
								}
							]
						}
					}
				}`
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		fmt.Fprintln(w, droplet)
	})
	mux.HandleFunc("/v2/account/keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(201)
			key := `{
				"ssh_key": {
					"id": 512189,
					"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
					"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
					"name": "Gofn"
				}
			}`
			fmt.Fprintln(w, key)
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			keys := `{
			"ssh_keys": [
				{
				"id": 512189,
				"fingerprint": "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa",
				"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAQQDDHr/jh2Jy4yALcK4JyWbVkPRaWmhck3IgCoeOO3z1e2dBowLh64QAM+Qb72pxekALga2oi4GvT+TlWNhzPH4V example",
				"name": "Gofn"
				}
			]
		}`
			fmt.Fprintln(w, keys)
		}
	})
	defineBrokenListSnapshotsEndpoint()
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Fatalf("Expected run with errors but  not has")
	}
}

func TestDeleteMachine(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/droplets/503/actions", func(w http.ResponseWriter, r *http.Request) {
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Expected parse request body without errors but has %q", err)
		}
		if strings.Contains(string(rBody), "shutdown") {
			w.WriteHeader(201)
			action := `{
				"action": {
					"id": 36077293,
					"status": "in-progress",
					"type": "shutdown",
					"started_at": "2014-11-04T17:08:03Z",
					"completed_at": null,
					"resource_id": 503,
					"resource_type": "droplet",
					"region": {"slug": "nyc3"}
				}
			}`
			fmt.Fprintln(w, action)
			return
		}
	})
	mux.HandleFunc("/v2/droplets/503", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	machine := &iaas.Machine{ID: "503"}
	do := &Digitalocean{}
	err := do.DeleteMachine(machine)
	if err != nil {
		t.Errorf("Expected run without errors but has %q", err)
	}
}

func TestDeleteMachineWithShutdownError(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/droplets/503/actions", func(w http.ResponseWriter, r *http.Request) {
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Expected parse request body without errors but has %q", err)
		}
		if strings.Contains(string(rBody), "shutdown") {
			w.WriteHeader(201)
			action := `{
				"action": {
					"id": 36077293,
					"status": "in-progress",
					"type": "shutdown",
					"started_at": "2014-11-04T17:08:03Z",
					"completed_at": null,
					"resource_id": 503,
					"resource_type": "droplet",
					"region": {"slug": "nyc3"},
				}
			}`
			fmt.Fprintln(w, action)
			return
		}
		if strings.Contains(string(rBody), "power_off") {
			w.WriteHeader(201)
			action := `{
				"action": {
					"id": 36077293,
					"status": "in-progress",
					"type": "power_off",
					"started_at": "2014-11-04T17:08:03Z",
					"completed_at": null,
					"resource_id": 503,
					"resource_type": "droplet",
					"region": {"slug": "nyc3"}
				}
			}`
			fmt.Fprintln(w, action)
			return
		}
	})
	mux.HandleFunc("/v2/droplets/503", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	machine := &iaas.Machine{ID: "503"}
	do := &Digitalocean{}
	err := do.DeleteMachine(machine)
	if err != nil {
		t.Errorf("Expected run without errors but has %q", err)
	}
}

func TestDeleteMachineWithShutdownErrorAndPowerOff(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/v2/droplets/503/actions", func(w http.ResponseWriter, r *http.Request) {
		rBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Expected parse request body without errors but has %q", err)
		}
		if strings.Contains(string(rBody), "shutdown") {
			w.WriteHeader(201)
			action := `{
				"action": {
					"id": 36077293,
					"status": "in-progress",
					"type": "shutdown",
					"started_at": "2014-11-04T17:08:03Z",
					"completed_at": null,
					"resource_id": 503,
					"resource_type": "droplet",
					"region": {"slug": "nyc3"},
				}
			}`
			fmt.Fprintln(w, action)
			return
		}
		if strings.Contains(string(rBody), "power_off") {
			w.WriteHeader(201)
			action := `{
				"action": {
					"id": 36077293,
					"status": "in-progress",
					"type": "power_off",
					"started_at": "2014-11-04T17:08:03Z",
					"completed_at": null,
					"resource_id": 503,
					"resource_type": "droplet",
					"region": {"slug": "nyc3"},
				}
			}`
			fmt.Fprintln(w, action)
			return
		}
	})
	mux.HandleFunc("/v2/droplets/503", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	machine := &iaas.Machine{ID: "503"}
	do := &Digitalocean{}
	err := do.DeleteMachine(machine)
	if err == nil {
		t.Errorf("expected run errors but not has %q", err)
	}
}

func TestDeleteMachineWrongAuth(t *testing.T) {
	os.Setenv("DIGITALOCEAN_API_URL", "://localhost")
	do := &Digitalocean{}
	err := do.DeleteMachine(&iaas.Machine{ID: "503"})
	if err == nil {
		t.Errorf("expected errors but run without errors")
	}
}

func TestCreateSnapshot(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets/123/actions", func(w http.ResponseWriter, r *http.Request) {
		action := `{
		"action": {
			"id": 36805022,
			"status": "in-progress",
			"type": "snapshot",
			"started_at": "2014-11-14T16:34:39Z",
			"completed_at": null,
			"resource_id": 3164450,
			"resource_type": "droplet",
			"region": {"slug": "nyc3"}
			}
		}`
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, action)

	})

	mac := &iaas.Machine{ID: "123"}
	do := &Digitalocean{}
	err := do.CreateSnashot(mac)
	if err != nil {
		t.Errorf("expected run without errors but has %q", err)
	}
}

func TestCreateSnapshotWrongAuth(t *testing.T) {
	os.Setenv("DIGITALOCEAN_API_URL", "://localhost")
	do := &Digitalocean{}
	err := do.CreateSnashot(&iaas.Machine{ID: "503"})
	if err == nil {
		t.Errorf("expected errors but run without errors")
	}
}

func TestCreateSnapshotActionError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets/123/actions", func(w http.ResponseWriter, r *http.Request) {
		action := `{
		"action": {
			"id": 36805022,
			"status": "in-progress",
			"type": "snapshot",
			"started_at": "2014-11-14T16:34:39Z",
			"completed_at": null,
			"resource_id": 3164450,
			"resource_type": "droplet",
			"region": {"slug": "nyc3"},
			}
		}`
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, action)

	})

	mac := &iaas.Machine{ID: "123"}
	do := &Digitalocean{}
	err := do.CreateSnashot(mac)
	if err == nil {
		t.Errorf("expected run with errors but not has")
	}
}
