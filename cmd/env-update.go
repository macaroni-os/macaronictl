/*
	Copyright © 2021-2022 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/

package cmd

import (
	"os"

	"github.com/funtoo/macaronictl/pkg/logger"
	"github.com/funtoo/macaronictl/pkg/portage"
	specs "github.com/funtoo/macaronictl/pkg/specs"

	"github.com/spf13/cobra"
)

func envUpdateCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "env-update",
		Aliases: []string{"eu"},
		Short:   "Updates environment settings automatically.",
		Long: `env-update reads the files in /etc/env.d and
automatically generates /etc/profile.env and /etc/ld.so.conf.
The ldconfig is run to update /etc/ld.so.cache after the
envs generation.
If you make changes to /etc/env.d, you should run env-update
yourself for changes to take effect immediately. Note that
this would only affect new processes. In order for  the changes
to affect your active shell, you will probably have to
run "source /etc/profile" first.

$> macaronictl env-update

`,
		PreRun: func(cmd *cobra.Command, args []string) {
			flags := cmd.Flags()

			config.Viper.BindPFlag("env-update.systemd", flags.Lookup("systemd"))
			config.Viper.BindPFlag("env-update.csh", flags.Lookup("csh"))
			config.Viper.BindPFlag("env-update.ldconfig", flags.Lookup("ldconfig"))
		},
		Run: func(cmd *cobra.Command, args []string) {

			log := logger.GetDefaultLogger()

			dryRun, _ := cmd.Flags().GetBool("dry-run")

			opts := portage.NewEnvUpdateOpts()
			opts.DryRun = dryRun
			opts.Csh = config.GetEnvUpdate().Csh
			opts.Systemd = config.GetEnvUpdate().Systemd
			opts.PrelinkCapable = config.GetEnvUpdate().Prelink
			opts.WithLdConfig = config.GetEnvUpdate().Ldconfig
			opts.Debug = config.GetGeneral().Debug

			err := portage.EnvUpdate("/", opts)
			if err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}

		},
	}

	flags := c.Flags()
	flags.Bool("dry-run", false, "Dry run commands.")
	flags.Bool("systemd", false, "Generate systemd environment file.")
	flags.Bool("csh", false, "Generate /etc/csh.env file.")
	flags.Bool("ldconfig", true, "Generate /etc/ld.so.cache and /etc/ld.so.conf.")

	return c
}
