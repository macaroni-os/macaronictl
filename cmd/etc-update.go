/*
	Copyright © 2021-2022 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	"os"
	"os/exec"

	"github.com/funtoo/macaronictl/pkg/logger"
	specs "github.com/funtoo/macaronictl/pkg/specs"

	"github.com/spf13/cobra"
)

func etcUpdateCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "etc-update",
		Aliases: []string{"etc"},
		Short:   "Handle configuration file updates.",
		Long: `At the moment it's a simple wrapper for Portage etc-update

$> macaronictl etc-update

`,
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()

			etcCommand := exec.Command("etc-update")
			etcCommand.Stdout = os.Stdout
			etcCommand.Stderr = os.Stderr
			etcCommand.Stdin = os.Stdin

			err := etcCommand.Start()
			if err != nil {
				log.Error("Error on start etc-update command: " + err.Error())
				os.Exit(1)
			}

			err = etcCommand.Wait()
			if err != nil {
				log.Error("Error on waiting etc-update command: " + err.Error())
				os.Exit(1)
			}

			os.Exit(etcCommand.ProcessState.ExitCode())
		},
	}

	return c
}
