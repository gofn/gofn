package digitalocean

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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
	do := &Digitalocean{}
	_, err := do.CreateMachine()
	if err == nil {
		t.Errorf("expected errors but run without errors")
	}
}
