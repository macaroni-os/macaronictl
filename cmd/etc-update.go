/*
	Copyright © 2021-2023 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	"os"

	"github.com/macaroni-os/macaronictl/pkg/logger"
	"github.com/macaroni-os/macaronictl/pkg/portage"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	"github.com/spf13/cobra"
)

func etcUpdateCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "etc-update",
		Aliases: []string{"etc"},
		Short:   "Handle configuration file updates.",
		Long: `handle configuration file updates and automatically
merge the CONFIG_PROTECT_MASK files.

$> macaronictl etc-update

`,
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.GetDefaultLogger()

			rootfs, _ := cmd.Flags().GetString("rootfs")
			paths, _ := cmd.Flags().GetStringArray("path")
			mpaths, _ := cmd.Flags().GetStringArray("mask-path")

			opts := portage.NewEtcUpdateOpts()
			opts.Paths = paths
			opts.MaskPaths = mpaths

			err := portage.EtcUpdate(rootfs, opts)
			if err != nil {
				log.Error("Error: " + err.Error())
				os.Exit(1)
			}
		},
	}

	flags := c.Flags()
	flags.String("rootfs", "/",
		"Override the default path where run etc-update. (experimental)")
	flags.StringArrayP("path", "p", []string{},
		"Scan one or more specific paths (CONFIG_PROTECT).")
	flags.StringArrayP("mask-path", "m", []string{},
		"Define one or more additional mask paths (CONFIG_PROTECT_MASK).")

	return c
}
