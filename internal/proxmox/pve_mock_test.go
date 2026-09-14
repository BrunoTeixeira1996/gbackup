package proxmox_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
)

// these hit a local httptest server instead of a real PVE host, by
// building PVE directly instead of calling Init()

func TestGetAllObjects_Mock_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/nodes/localhost/lxc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"vmid":100,"name":"lxc1","status":"running"},{"vmid":101,"name":"lxc2","status":"stopped"}]}`)
	})
	mux.HandleFunc("/nodes/localhost/qemu", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"vmid":200,"name":"vm1","status":"running"}]}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	pve := proxmox.PVE{
		API: proxmox.Boilerplate{
			Url:           srv.URL,
			Node:          "localhost",
			Authorization: "test",
		},
	}

	if err := pve.GetAllObjects(); err != nil {
		t.Fatalf("GetAllObjects() returned unexpected error: %v", err)
	}

	if len(pve.LXCs) != 2 {
		t.Fatalf("len(LXCs) = %d, want 2", len(pve.LXCs))
	}
	if len(pve.VMs) != 1 {
		t.Fatalf("len(VMs) = %d, want 1", len(pve.VMs))
	}
	if pve.LXCs[0].Name != "lxc1" {
		t.Errorf("LXCs[0].Name = %q, want %q", pve.LXCs[0].Name, "lxc1")
	}
	if pve.VMs[0].Name != "vm1" {
		t.Errorf("VMs[0].Name = %q, want %q", pve.VMs[0].Name, "vm1")
	}
}

func TestGetAllObjects_Mock_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	pve := proxmox.PVE{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	if err := pve.GetAllObjects(); err == nil {
		t.Fatal("expected an error when the API returns a non-200 status, got nil")
	}
}

func TestGetAllObjects_Mock_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not valid json`)
	}))
	defer srv.Close()

	pve := proxmox.PVE{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	if err := pve.GetAllObjects(); err == nil {
		t.Fatal("expected an error when the API returns malformed JSON, got nil")
	}
}

func TestGetAllObjects_Mock_UnreachableHost(t *testing.T) {
	pve := proxmox.PVE{
		API: proxmox.Boilerplate{
			Url:  "http://127.0.0.1:1", // nothing listens here
			Node: "localhost",
		},
	}

	if err := pve.GetAllObjects(); err == nil {
		t.Fatal("expected an error when the API host is unreachable, got nil")
	}
}
