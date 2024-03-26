/*
Copyright © 2020-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/logger"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2020-2024 Macaroni OS - Daniele Rondina

Macaroni OS System Management Tool`

	MACARONICTL_VERSION = `0.10.0`
)

var (
	BuildTime   string
	BuildCommit string
)

func initConfig(config *specs.MacaroniCtlConfig) {
	// Set env variable
	config.Viper.SetEnvPrefix(specs.MACARONICTL_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(specs.MACARONICTL_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)
}

func initCommand(rootCmd *cobra.Command, config *specs.MacaroniCtlConfig) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "Macaronictl configuration file")
	pflags.BoolP("debug", "d", config.Viper.GetBool("general.debug"),
		"Enable debug output.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("general.debug", pflags.Lookup("debug"))

	rootCmd.AddCommand(
		envUpdateCommand(config),
		etcUpdateCommand(config),
		kernelCmdCommand(config),
		browserCmdCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.MacaroniCtlConfig = specs.NewMacaroniCtlConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      fmt.Sprintf("%s-g%s %s", MACARONICTL_VERSION, BuildCommit, BuildTime),
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			v.SetConfigType("yml")
			if v.Get("config") == "" {
				config.Viper.AddConfigPath(".")
			} else {
				v.SetConfigFile(v.Get("config").(string))
			}

			// Parse configuration file
			err = config.Unmarshal()
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					// Config file not found; ignore error if desired
				} else {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			// Initialize logger
			log := logger.NewMacaroniCtlLogger(config)
			log.SetAsDefault()
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
