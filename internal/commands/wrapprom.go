package commands

import (
	"context"
	"flag"
	"fmt"
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
		exitCode   int
	)

	ctx := context.Background()

	// executes the given command
	start := func(ctx context.Context, args []string) (io.Reader, error) {
		c = exec.CommandContext(ctx, name, command...)
		c.Stderr = os.Stderr
		rc, err := c.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("[prom error] could not exec.CommandContext %s\n", err)
		}
		stdoutPipe = rc

		log.Printf("[prom info] executing: %q\n", c.Args)
		if err := c.Start(); err != nil {
			return nil, fmt.Errorf("[prom error] could not c.Start %s\n", err)
		}
		return rc, nil
	}
	// waits for the exit code
	wait := func() int {
		defer stdoutPipe.Close()
		if err := c.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
					return exitCode
				}
			}
			log.Printf("[prom error] error while waiting: %s\n", err)
			exitCode = 1
			return exitCode
		}
		exitCode = 0
		return exitCode
	}

	params := rsyncprom.WrapParams{
		Pushgateway: pg,
		Instance:    instance,
		Job:         commandType, // this is toExternal or toNAS
	}
	// executes WrapRsync from rsyncprom and export metrics to prometheus.
	// WrapRsync only returns an error if starting the command or parsing
	// its output failed - it records the exit code as a metric but never
	// as a returned error, so a real rsync failure (e.g. protocol
	// incompatibility) would otherwise be reported as success.
	err = rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)
	if err == nil && exitCode != 0 {
		err = fmt.Errorf("[prom error] rsync exited with code %d", exitCode)
	}
	log.Printf("[prom info] executing %s %s -> result: %s\n", instance, params.Job,
		func() string {
			if err == nil {
				return "OK ✅"
			}
			return fmt.Sprintf("ERROR ❌: %s", err.Error())
		}(),
	)

	/*	switch commandType {
		case "toExternal":
			params := rsyncprom.WrapParams{
				Pushgateway: pg,
				Instance:    instance,
				Job:         "toExternal",
			}
			// executes WrapRsync from rsyncprom and export metrics to prometheus
			err = rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)
			log.Printf("[prom info] executing %s %s -> result: %s\n", instance, params.Job,
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
			log.Printf("[prom info] executing %s %s -> result: %s\n", instance, params.Job,
				func() string {
					if err == nil {
						return "OK"
					}
					return err.Error()
				}(),
			)
		}
	*/
	return err
}
