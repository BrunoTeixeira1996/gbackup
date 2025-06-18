package commands

import (
	"strings"
)

// Used like this to be able to mock the backup in tests
var RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
	rcommand := strings.Fields(cmd)
	if err := ExecCmdToProm("rsync", rcommand, to, target, pushgatewayURL); err != nil {
		return err
	}
	return nil
}
