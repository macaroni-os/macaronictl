/*
	Copyright © 2021 RockHopper OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	cmdkernel "github.com/funtoo/rhctl/cmd/kernel"
	specs "github.com/funtoo/rhctl/pkg/specs"
	"github.com/spf13/cobra"
)

func kernelCmdCommand(config *specs.RhCtlConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "kernel",
		Short: "Manage system kernels and initrd.",
		Long:  `Manage kernels and initrd images of your system.`,
	}

	cmd.AddCommand(
		cmdkernel.NewListcommand(config),
		cmdkernel.NewGeninitrdCommand(config),
		cmdkernel.NewProfilesCommand(config),
	)

	return cmd
}
