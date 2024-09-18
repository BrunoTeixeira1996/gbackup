package proxmox

type Proxmox struct {
	TokenID       string
	Secret        string
	APIUrl        string
	Node          string
	Authorization string
}

// TODO
func (p *Proxmox) Init() error {
	return nil
}

// TODO
func getCurrentLXC() error {
	return nil
}

// TODO
func getCurrentVMs() error {
	return nil
}
