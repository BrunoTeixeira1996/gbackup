package commands

import "strings"

func RsyncCommand(cmd, to, target, pushgatewayURL string) error {
	rcommand := strings.Fields(cmd)
	if err := ExecCmdToProm("rsync", rcommand, to, target, pushgatewayURL); err != nil {
		return err
	}
	return nil
}
