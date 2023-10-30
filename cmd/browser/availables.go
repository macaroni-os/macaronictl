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

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewAvailableCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "available",
		Aliases: []string{"availables", "a"},
		Short:   "List available browsers and their customization.",
		Long: `Shows browsers available in configured repositories.

$ macaronictl browser availables

NOTE: It works only if the repositories are synced.
`,
		PreRun: func(cmd *cobra.Command, args []string) {
			systemDir, _ := cmd.Flags().GetString("system-dir")
			homeDir, _ := cmd.Flags().GetString("home-dir")

			if systemDir == "" {
				fmt.Println("Invalid system-dir option.")
				os.Exit(1)
			}

			if homeDir == "" {
				fmt.Println("Invalid home-dir option.")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {

			//log := logger.GetDefaultLogger()
			jsonOutput, _ := cmd.Flags().GetBool("json")
			catalogFile, _ := cmd.Flags().GetString("catalog-file")
			systemDir, _ := cmd.Flags().GetString("system-dir")
			homeDir, _ := cmd.Flags().GetString("home-dir")

			stones, err := browser.AvailableBrowsers(false, config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			stonesInstalled, err := browser.AvailableBrowsers(true, config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
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

			if !jsonOutput {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Package",
					"Package Version",
					"System Options",
					"User Options",
					"Engine",
					"Binaries",
				})

				installedStonesMap := stonesInstalled.ToMap()

				// Create response struct
				for _, s := range stones.Stones {

					withSystemOpts := false
					withHomeOpts := false
					binaries := "N/A"

					engine, pkg := configurator.GetSystemConfig().GetEngineAndPackage(s.GetName())
					if !configurator.GetSystemConfig().IsEmpty() && pkg != nil &&
						pkg.HasOptions() {
						withSystemOpts = true
						binaries = pkg.Binary
					}

					homePkg := configurator.GetHomeConfig().GetPackage(s.GetName())
					if !configurator.GetHomeConfig().IsEmpty() && homePkg != nil {
						withHomeOpts = true
						binaries = homePkg.Binary
					}

					engineName := "N/A"
					if engine != nil {
						engineName = engine.Name
					} else {
						// POST: Get engine from catalog.
						engine, pkg = configurator.GetCatalog().GetEngineAndPackage(s.GetName())
						if engine != nil {
							engineName = engine.Name
						}

						if pkg != nil {
							binaries = pkg.Binary
						}
					}

					name := s.GetName()
					version := "v" + s.GetVersion()

					if _, ok := installedStonesMap.Stones[s.GetName()]; ok {
						name = fmt.Sprintf("%s", aurora.Bold(name))
						version = fmt.Sprintf("%s", aurora.Bold(version))
					}

					row := []string{
						name,
						version,
						fmt.Sprintf("%v", withSystemOpts),
						fmt.Sprintf("%v", withHomeOpts),
						engineName,
						binaries,
					}
					table.Append(row)
				}

				table.Render()
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
	flags.Bool("json", false, "JSON output")
	flags.String("catalog-file", "/usr/share/macaroni/browsers/catalog",
		"Specify the directory of the catalog file of all engines options.")
	flags.String("system-dir", "/etc/macaroni/browsers",
		"Override the directory of the system configuration with engines options.")
	flags.String("home-dir", browserConfigsHomedir,
		"Override the directory of the user with engines options.")
	return c
}
