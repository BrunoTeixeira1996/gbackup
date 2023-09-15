package internal

import (
	"bytes"
	"os"

	"golang.org/x/crypto/ssh"
)

// Function that reads an OpenSSH key and provides it as a ssh.ClientAuth.
func openSSHClientAuth(path string) (ssh.AuthMethod, error) {
	privateKey, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	return ssh.PublicKeys(signer), err
}

// Function that returns a ssh connection
func newSshConnection(host, keypath string) (*ssh.Client, error) {
	clientauth, err := openSSHClientAuth(keypath)
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{clientauth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host+":22", clientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Function executes a given cmd in a ssh connection
func ExecuteCmdSSH(cmd, host string, keypath string) error {
	conn, err := newSshConnection(host, keypath)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return nil
}
