package commands

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/stapelberg/rsyncprom"
)

// stolen from https://github.com/stapelberg/rsyncprom
// Function that executes command and gather metrics to prometheus
// i've modified a little bit this to make it usefull for my use case
func ExecCmdToProm(name string, command []string, commandType string, instance string, pg string) error {
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

		log.Printf("[prom] executing: %q\n", c.Args)
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
			log.Printf("[prom] error while waiting: %v\n", err)
			return 1
		}
		return 0
	}

	switch commandType {
	case "toExternal":
		params := rsyncprom.WrapParams{
			Pushgateway: pg,
			Instance:    instance,
			Job:         "toExternal",
		}
		// executes WrapRsync from rsyncprom and export metrics to prometheus
		err = rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)
		log.Printf("[prom] executing %s %s -> result: %s\n", instance, params.Job,
			func() string {
				if err == nil {
					return "OK"
				}
				return err.Error()
			}(),
		)

	case "toNAS":
		params := rsyncprom.WrapParams{
			Pushgateway: pg,
			Instance:    instance,
			Job:         "toNAS",
		}
		// executes WrapRsync from rsyncprom and export metrics to prometheus
		err = rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)
		log.Printf("[prom] executing %s %s -> result: %s\n", instance, params.Job,
			func() string {
				if err == nil {
					return "OK"
				}
				return err.Error()
			}(),
		)
	}

	return err
}
