package targets

import (
	"bytes"
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/stapelberg/rsyncprom"
	"golang.org/x/crypto/ssh"
)

// Function that reads an OpenSSH key and provides it as a ssh.ClientAuth.
func openSSHClientAuth(path string) (ssh.AuthMethod, error) {
	privateKey, err := ioutil.ReadFile(path)
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
func executeCmdSSH(cmd, host string, keypath string) error {
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

// stolen from https://github.com/stapelberg/rsyncprom
// Function that executes command and gather and gather metrics to prometheus
// i've modified a little bit this to make it usefull for my use case
func execCmdToProm(name string, command []string, commandType string, instance string) error {
	var (
		c          *exec.Cmd
		stdoutPipe io.ReadCloser
		err        error
	)

	ctx := context.Background()

	// executes the given command
	start := func(ctx context.Context, args []string) (io.Reader, error) {
		c = exec.CommandContext(ctx, name, command...)
		c.Stderr = os.Stderr
		rc, err := c.StdoutPipe()
		if err != nil {
			return nil, err
		}
		stdoutPipe = rc

		log.Printf("Starting %s: %q\n", commandType, c.Args)
		if err := c.Start(); err != nil {
			return nil, err
		}
		return rc, nil
	}
	// waits for the exit code
	wait := func() int {
		defer stdoutPipe.Close()
		if err := c.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					code := status.ExitStatus()
					return code
				}
			}
			log.Print(err)
			return 1
		}
		return 0
	}

	switch commandType {
	case "rsync":
		params := rsyncprom.WrapParams{
			Pushgateway: "http://192.168.30.77:9091",
			Instance:    instance,
			Job:         "rsync",
		}
		// executes WrapRsync from rsyncprom and export metrics to prometheus
		err = rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)

	case "cmd":
		params := rsyncprom.WrapParams{
			Pushgateway: "http://192.168.30.77:9091",
			Instance:    instance,
			Job:         "cmd",
		}
		// executes wrampCmd and export metrics to prometheus
		err = wrapCmd(ctx, &params, flag.Args(), start, wait)
	}

	return err
}

// TODO
// write log to a log file maybe?
// the only thing that matters here is the start and end time and then
// the exit code of the command
// stolen from https://github.com/stapelberg/rsyncprom/blob/main/rsyncprom.go
func wrapCmd(ctx context.Context, params *rsyncprom.WrapParams, args []string, start func(context.Context, []string) (io.Reader, error), wait func() int) error {
	log.Printf("push gateway: %q", params.Pushgateway)

	startTimeMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: params.Job + "_start_timestamp_seconds",
		Help: "The timestamp of the cmd start",
	})
	startTimeMetric.SetToCurrentTime()
	pushAll := func(collectors ...prometheus.Collector) {
		p := push.New(params.Pushgateway, params.Job).
			Grouping("instance", params.Instance)
		for _, c := range collectors {
			p.Collector(c)
		}
		if err := p.Add(); err != nil {
			log.Print(err)
		}
	}
	pushAll(startTimeMetric)

	exitCode := 0

	// defer will wait for the wait() function to finish
	defer func() {
		log.Printf("Pushing exit code %d", exitCode)
		exitCodeMetric := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: params.Job + "_exit_code",
			Help: "The exit code (0 = success, non-zero = failure)",
		})
		exitCodeMetric.Set(float64(exitCode))
		// end timestamp is push_time_seconds
		pushAll(exitCodeMetric)
	}()

	_, err := start(ctx, args)
	if err != nil {
		return err
	}

	log.Printf("Parsing cmd output")

	exitCode = wait()

	return nil
}
