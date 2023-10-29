/*
	Copyright © 2021-2023 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	cmdbrowser "github.com/macaroni-os/macaronictl/cmd/browser"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/spf13/cobra"
)

func browserCmdCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "browser",
		Aliases: []string{"b"},
		Short:   "Manage browsers bootstrap options.",
		Long:    `Manage browsers binaries and their default bootstrap options.`,
	}

	cmd.AddCommand(
		cmdbrowser.NewAvailableCommand(config),
		cmdbrowser.NewConfigureCommand(config),
	)

	return cmd
}
