package proxmox

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Boilerplate struct {
	TokenID       string
	Secret        string
	Url           string
	Node          string
	Authorization string
}

type PVE struct {
	API  Boilerplate
	VMs  []Data
	LXCs []Data
}

// Data from LXC and VM
type Data struct {
	Type      string  `json:"type"`
	Swap      int64   `json:"swap"`
	PID       int     `json:"pid"`
	MaxSwap   int64   `json:"maxswap"`
	DiskWrite int64   `json:"diskwrite"`
	CPUs      int     `json:"cpus"`
	Tags      string  `json:"tags"`
	NetIn     int64   `json:"netin"`
	CPU       float64 `json:"cpu"`
	Mem       int64   `json:"mem"`
	NetOut    int64   `json:"netout"`
	VMID      int     `json:"vmid"`
	Disk      int64   `json:"disk"`
	Uptime    int64   `json:"uptime"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	DiskRead  int64   `json:"diskread"`
	MaxMem    int64   `json:"maxmem"`
	MaxDisk   int64   `json:"maxdisk"`
}

// An object can be a LXC or a VM
type Object struct {
	Data []Data `json:"data"`
}

// Boilerplate function to perform an http request to API
func (b *Boilerplate) request(rType, apiPath string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	requestBody := fmt.Sprintf("%s/nodes/%s/%s", b.Url, b.Node, apiPath)

	req, err := http.NewRequest(rType, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("[boilerplate error] could not create new request: %s\n", err)
	}

	req.Header.Set("Authorization", b.Authorization)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[boilerplate error] could not perform client.Do: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[boilerplate error] request to %s returned status code: %d\n", requestBody, resp.StatusCode)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[boilerplate error] could not read response body: %s\n", err)
	}

	return res, nil
}

func (p *PVE) Init() error {
	tokenID := os.Getenv("PVE_TOKENID")
	secret := os.Getenv("PVE_SECRET")

	if tokenID == "" || secret == "" {
		return fmt.Errorf("[pve error] please provide the PVE token and secret env vars\n")
	}

	p.API.TokenID = tokenID
	p.API.Secret = secret
	p.API.Url = "https://192.168.30.3:8006/api2/json"
	p.API.Node = "localhost"
	p.API.Authorization = fmt.Sprintf("PVEAPIToken=%s=%s", p.API.TokenID, p.API.Secret)

	return nil
}

// Assigns LXC and VMs into PVE struct
func (p *PVE) getAllObjects() error {
	var (
		tL  Object
		tV  Object
		err error
	)

	if tL, err = getEntriesForObject(*p, "lxc"); err != nil {
		return fmt.Errorf("[getAllObjects] could not get all for LXC: %s\n", err)
	}
	p.LXCs = tL.Data

	if tV, err = getEntriesForObject(*p, "qemu"); err != nil {
		return fmt.Errorf("[getAllObjects] could not get all for VM: %s\n", err)
	}
	p.VMs = tV.Data

	return nil
}

// Gets all entries for an object (lxc or vm)
func getEntriesForObject(pve PVE, objType string) (Object, error) {
	var (
		response []byte
		err      error
		ob       Object
	)

	if response, err = pve.API.request("GET", objType); err != nil {
		return Object{}, fmt.Errorf("[getEntriesForObject] could not get containers from API (%s): %s\n", objType, err)
	}

	if err = json.Unmarshal(response, &ob); err != nil {
		return Object{}, fmt.Errorf("[getEntriesForObject] could not unmarshal response: %s\n", err)
	}

	return ob, nil
}

/*func Xest() {
	pve := &PVE{}
	var err error

	if err = pve.Init(); err != nil {
		log.Println(err)
	}

	if err = pve.getAllObjects(); err != nil {
		log.Println(err)
	}

	log.Println("TOTAL:", len(pve.LXCs)+len(pve.VMs))
}
*/
