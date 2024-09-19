package nas

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
)

// Check if SSH is open
func checkSSH(ctx context.Context, addr string) error {
	// poll every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// set a timeout for the entire polling operation - 5 minutes
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// the context's timeout or cancellation was triggered
			return fmt.Errorf("[nas error] context ended before SSH became reachable on %s: %v", addr, ctx.Err())
		case <-ticker.C:
			// attempt to connect to TCP port 22 (SSH)
			conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr+":22")
			if err == nil {
				defer conn.Close()
				// SSH is reachable
				log.Printf("[nas info] ssh is open on %s\n", addr)
				return nil
			}
			log.Printf("[nas info] ssh not reachable on %s, retrying...\n", addr)
		}
	}
}

// Sends magic packet WoL
func sendMagicPacket(nasMac string) error {
	hwaddr, err := net.ParseMAC(nasMac)
	if err != nil {
		return fmt.Errorf("[nas error] could not parse mac %s: %s\n", nasMac, err)
	}
	if got, want := len(hwaddr), 6; got != want {
		return fmt.Errorf("[nas error] could not send magic packet unexpected number of parts in hardware address %q: got %d, want %d", nasMac, got, want)
	}

	socket, err := net.DialUDP("udp4",
		nil,
		&net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: 9, // discard
		})
	if err != nil {
		return fmt.Errorf("DialUDP(broadcast:discard): %v", err)
	}
	// https://en.wikipedia.org/wiki/Wake-on-LAN#Magic_packet
	payload := append(bytes.Repeat([]byte{0xff}, 6), bytes.Repeat(hwaddr, 16)...)
	if _, err := socket.Write(payload); err != nil {
		return fmt.Errorf("[nas error] could not append magic byte: %s\n", err)
	}
	return socket.Close()
}

// Wakes up the NAS
func Wakeup(nas config.NAS, ctx context.Context) error {
	log.Printf("[nas info] validate if nas is reachable\n")
	if isReachable(nas.IP) {
		log.Printf("[nas info] %s (%s) is up ... ignoring sending magic packet", nas.Name, nas.IP)
		return nil
	}

	log.Printf("[nas info] sending magic packet to %s (%s-%s)\n", nas.Name, nas.IP, nas.MAC)
	if err := sendMagicPacket(nas.MAC); err != nil {
		return err
	}
	log.Printf("[nas info] magic packet sent to %s (%s-%s)\n", nas.Name, nas.IP, nas.MAC)

	{
		ctx, canc := context.WithTimeout(ctx, 5*time.Minute)
		defer canc()
		// check if port 22 is open already
		log.Printf("[nas info] checking if ssh is open on %s\n", nas.Name)
		if err := checkSSH(ctx, nas.IP); err != nil {
			return err
		}
		log.Printf("[nas info] host %s is now awake\n", nas.Name)
	}

	return nil
}

// Check if the NAS is reachable
func isReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr+":22", 5*time.Second)
	if err != nil {
		log.Printf("[nas info] connection failed, likely system is down\n")
		return false
	}
	conn.Close()
	return true
}

// Shuts down the NAS
func Shutdown(nas config.NAS) error {
	// check if the system is reachable before issuing the shutdown command
	if !isReachable(nas.IP) {
		log.Printf("[nas info] %s (%s) is already down\n", nas.Name, nas.IP)
		return nil
	}

	cmd := exec.Command("ssh", nas.Name, "sudo", "shutdown", "-P", "0")
	if err := cmd.Run(); err != nil {
		return err
	}

	time.Sleep(20 * time.Second)

	if !isReachable(nas.IP) {
		log.Printf("[nas info] confirmed that %s (%s) is down\n", nas.Name, nas.IP)
	}

	return nil
}
