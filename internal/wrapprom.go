package internal

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
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
			Pushgateway: pg,
			Instance:    instance,
			Job:         "rsync",
		}
		// executes WrapRsync from rsyncprom and export metrics to prometheus
		err = rsyncprom.WrapRsync(ctx, &params, flag.Args(), start, wait)

	case "cmd":
		params := rsyncprom.WrapParams{
			Pushgateway: pg,
			Instance:    instance,
			Job:         "cmd",
		}
		// executes wrampCmd and export metrics to prometheus
		err = wrapCmd(ctx, &params, flag.Args(), start, wait)
	}

	return err
}

// stolen from https://github.com/stapelberg/rsyncprom/blob/main/rsyncprom.go
// Function that wraps unix commands collecting metrics to prometheus
// the only thing that matters here is the start and end time and then
// the exit code of the command
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
