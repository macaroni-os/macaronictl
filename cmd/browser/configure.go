/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmdbrowser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/macaroni-os/macaronictl/pkg/browser"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"

	"github.com/spf13/cobra"
)

func NewConfigureCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "configure [pkg]",
		Aliases: []string{"conf", "c"},
		Short:   "Configure bootstrap options of a specific browser.",
		Long: `Shows browsers available in configured repositories.

$ macaronictl browser conf www-client/brave-bin --system --defaults

$ macaronictl browser conf www-client/brave-bin --user --defaults

$ macaronictl browser conf www-client/brave-bin --user --without-opts

NOTE: It works only if the repositories are synced.
`,
		PreRun: func(cmd *cobra.Command, args []string) {
			systemDir, _ := cmd.Flags().GetString("system-dir")
			homeDir, _ := cmd.Flags().GetString("home-dir")
			defaults, _ := cmd.Flags().GetBool("defaults")
			withoutOpts, _ := cmd.Flags().GetBool("without-opts")
			system, _ := cmd.Flags().GetBool("system")
			user, _ := cmd.Flags().GetBool("user")

			if systemDir == "" {
				fmt.Println("Invalid system-dir option.")
				os.Exit(1)
			}

			if homeDir == "" {
				fmt.Println("Invalid home-dir option.")
				os.Exit(1)
			}

			if !defaults && !withoutOpts {
				fmt.Println("Use --defaults or --without-opts")
				os.Exit(1)
			}

			if defaults && withoutOpts {
				fmt.Println("Both options --defaults and --without-opts set.")
				os.Exit(1)
			}

			if system && user {
				fmt.Println("Use --system or --user")
				os.Exit(1)
			}

			if len(args) == 0 {
				fmt.Println("Package name mandatory.")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {

			//log := logger.GetDefaultLogger()
			catalogFile, _ := cmd.Flags().GetString("catalog-file")
			systemDir, _ := cmd.Flags().GetString("system-dir")
			homeDir, _ := cmd.Flags().GetString("home-dir")
			system, _ := cmd.Flags().GetBool("system")
			user, _ := cmd.Flags().GetBool("user")
			//binary, _ := cmd.Flags().GetBool("exec")
			defaults, _ := cmd.Flags().GetBool("defaults")
			//withoutOpts, _ := cmd.Flags().GetBool("without-opts")

			pkgname := args[0]

			if !system && !user {
				user = true
			}

			opts := &browser.BrowserConfiguratorOpts{
				Systemdir:   systemDir,
				Homedir:     homeDir,
				Catalogfile: catalogFile,
			}

			// Create configurator
			configurator, err := browser.NewBrowserConfigurator(opts)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			err = configurator.WorkingOnPackage(pkgname, system)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if defaults {

				// Retrieve defaults options from catalog
				catEngine := configurator.Catalog.GetEngine(configurator.Engine.Name)
				if catEngine == nil {
					fmt.Println(fmt.Sprintf("Engine %s not found on catalog.",
						configurator.Engine.Name,
					))
					os.Exit(1)
				}

				defOpts := catEngine.GetDefaultOptions()
				if len(defOpts) == 0 {
					fmt.Println("No defaults options availables. Nothing to do.")
					os.Exit(0)
				}

				configurator.Package.SetEnabledOptions(defOpts)

			} else {
				// POST: without-ops is set
				configurator.Package.ClearOptions()
			}

			var f string
			if system {
				f = filepath.Join(systemDir, configurator.Engine.Name+".yml")
			} else {
				f = filepath.Join(homeDir, configurator.Engine.Name+".yml")
			}

			err = configurator.Engine.WriteConfig(f)
			if err != nil {
				fmt.Println("ERROR On write engine file:", err.Error())
				os.Exit(1)
			}
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	browserConfigsHomedir := filepath.Join(homeDir, ".local/share/macaroni/browsers")

	flags := c.Flags()
	flags.String("catalog-file", "/usr/share/macaroni/browsers/catalog",
		"Specify the directory of the catalog file of all engines options.")
	flags.String("system-dir", "/etc/macaroni/browsers",
		"Override the directory of the system configuration with engines options.")
	flags.String("home-dir", browserConfigsHomedir,
		"Override the directory of the user with engines options.")
	flags.Bool("user", false, "Set bootstrap option for user.")
	flags.Bool("system", false, "Set bootstrap option on system. Need root permissions.")
	flags.Bool("exec", false, "Update script of the binary. Need root permissions.")
	flags.Bool("defaults", false, "Set catalog defaults options to specified package.")
	flags.Bool("without-opts", false, "Disable all options to specified package.")
	flags.Bool("purge", false, "Remove system option from system. Need root permissions.")

	return c
}
