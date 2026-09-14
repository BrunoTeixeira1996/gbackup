package nas

import "testing"

// only testing the invalid MAC path, valid one sleeps 10s and sends a real packet
func TestSendMagicPacket_InvalidMAC(t *testing.T) {
	if err := sendMagicPacket("not-a-valid-mac"); err == nil {
		t.Error("sendMagicPacket() with an invalid MAC = nil error, want an error")
	}
}

func TestSendMagicPacket_EmptyMAC(t *testing.T) {
	if err := sendMagicPacket(""); err == nil {
		t.Error("sendMagicPacket() with an empty MAC = nil error, want an error")
	}
}
