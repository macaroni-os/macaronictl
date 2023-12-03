/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmdbrowser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/macaroni-os/macaronictl/pkg/browser"
	"github.com/macaroni-os/macaronictl/pkg/logger"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"

	"github.com/spf13/cobra"
)

func NewConfigureCommand(config *specs.MacaroniCtlConfig) *cobra.Command {
	c := &cobra.Command{
		Use:     "configure [pkg]",
		Aliases: []string{"conf", "c"},
		Short:   "Configure bootstrap options of a specific browser.",
		Long: `Shows browsers available in configured repositories.

# Generate the system yaml file with the default options from catalog.
$> macaronictl browser conf www-client/brave-bin --system --defaults

# Generate the user yaml file with the default options from catalog.
$> macaronictl browser conf www-client/brave-bin --user --defaults

# Generate the user yaml file without options for the selected package.
$> macaronictl browser conf www-client/brave-bin --user --without-opts

# Generate the user yaml file awithout options and the user include file
# for the selected package.
$> macaronictl browser conf www-client/brave-bin --user --without-opts --exec

# Generate the binary script of the package and the system includes scripts.
# Normally, this command is executed on package finalizer.
$> macaronictl browser conf www-client/brave-bin --exec --system  --defaults

# Generate the user include and YAML files with the default options
$> macaronictl browser conf www-client/brave-bin --exec --user  --defaults

# Remove the user include file.
$> macaronictl browser conf www-client/brave-bin --purge --user

# Remove the system include file and the binary of the package
$> macaronictl browser conf www-client/brave-bin --purge --system

# Update the user include file. Normally, used when the user YAML file
# is been modified manually.
$> macaronictl browser conf www-client/brave-bin --user --only-update-includes

# Update the system include file. Normally, used when the user YAML file
# is been modified manually.
$> macaronictl browser conf www-client/brave-bin --system --only-update-includes

NOTE: It works only if the repositories are synced.
`,
		PreRun: func(cmd *cobra.Command, args []string) {
			systemDir, _ := cmd.Flags().GetString("system-dir")
			homeDir, _ := cmd.Flags().GetString("home-dir")
			defaults, _ := cmd.Flags().GetBool("defaults")
			withoutOpts, _ := cmd.Flags().GetBool("without-opts")
			system, _ := cmd.Flags().GetBool("system")
			user, _ := cmd.Flags().GetBool("user")
			binary, _ := cmd.Flags().GetBool("exec")
			purge, _ := cmd.Flags().GetBool("purge")
			onlyUpdateIncludes, _ := cmd.Flags().GetBool("only-update-includes")

			if systemDir == "" {
				fmt.Println("Invalid system-dir option.")
				os.Exit(1)
			}

			if homeDir == "" {
				fmt.Println("Invalid home-dir option.")
				os.Exit(1)
			}

			if !defaults && !withoutOpts && !purge && !onlyUpdateIncludes {
				fmt.Println("Use --defaults or --without-opts")
				os.Exit(1)
			}

			if defaults && withoutOpts {
				fmt.Println("Both options --defaults and --without-opts set.")
				os.Exit(1)
			}

			if purge && binary {
				fmt.Println("Both --purge and --exec not usable together.")
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

			log := logger.GetDefaultLogger()
			catalogFile, _ := cmd.Flags().GetString("catalog-file")
			systemDir, _ := cmd.Flags().GetString("system-dir")
			homeDir, _ := cmd.Flags().GetString("home-dir")
			system, _ := cmd.Flags().GetBool("system")
			user, _ := cmd.Flags().GetBool("user")
			binary, _ := cmd.Flags().GetBool("exec")
			purge, _ := cmd.Flags().GetBool("purge")
			onlyUpdateIncludes, _ := cmd.Flags().GetBool("only-update-includes")
			defaults, _ := cmd.Flags().GetBool("defaults")
			withoutOpts, _ := cmd.Flags().GetBool("without-opts")

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
				log.Fatal(err.Error())
			}

			err = configurator.WorkingOnPackage(pkgname, system)
			if err != nil {
				log.Fatal(err.Error())
			}

			if defaults && !purge {

				// Retrieve defaults options from catalog
				catEngine := configurator.Catalog.GetEngine(configurator.Engine.Name)
				if catEngine == nil {
					log.Fatal(fmt.Sprintf("Engine %s not found on catalog.",
						configurator.Engine.Name,
					))
				}

				if !onlyUpdateIncludes {
					defOpts := catEngine.GetDefaultOptions()
					if len(defOpts) == 0 {
						fmt.Println("No defaults options availables. Nothing to do.")
						os.Exit(0)
					}

					configurator.Package.SetEnabledOptions(defOpts)
				}

			} else if withoutOpts && !onlyUpdateIncludes {
				// POST: without-ops is set
				configurator.Package.ClearOptions()
			}

			if purge {

				if system {
					// Write system include
					systemInclude := filepath.Join(systemDir,
						fmt.Sprintf("%s.%s.inc", configurator.Engine.Name,
							filepath.Base(configurator.Package.Binary)))

					if utils.Exists(systemInclude) {
						err = os.Remove(systemInclude)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error on removing file %s: %s",
								systemInclude, err.Error()))
						}
						log.Info(fmt.Sprintf("Removed file %s", systemInclude))
					}

					if utils.Exists(configurator.Package.Binary) {
						err = os.Remove(configurator.Package.Binary)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error on removing file %s: %s",
								configurator.Package.Binary, err.Error()))
						}
						log.Info(fmt.Sprintf("Removed file %s", configurator.Package.Binary))
					}

					engine, _ := configurator.GetSystemConfig().GetEngineAndPackage(pkgname)
					if engine != nil && engine.NumPackages() == 1 {
						// POST: The engine configured contains only the purged package.
						f := filepath.Join(systemDir, configurator.Engine.Name+".yml")

						if utils.Exists(f) {
							err = os.Remove(f)
							if err != nil {
								log.Fatal(fmt.Sprintf("Error on removing file %s: %s",
									f, err.Error()))
							}
							log.Info(fmt.Sprintf("Removed file %s", f))
						}
					}
				} else {
					// Write user include
					homeInclude := filepath.Join(homeDir,
						fmt.Sprintf("%s.%s.inc", configurator.Engine.Name,
							filepath.Base(configurator.Package.Binary)))

					if utils.Exists(homeInclude) {
						err = os.Remove(homeInclude)
						if err != nil {
							log.Fatal(fmt.Sprintf("Error on removing file %s: %s",
								homeInclude, err.Error()))
						}
						log.Info(fmt.Sprintf("Removed file %s", homeInclude))
					}
				}

			} else {

				var f string
				if system {
					f = filepath.Join(systemDir, configurator.Engine.Name+".yml")
					log.InfoC(
						fmt.Sprintf(`Configuring browser start flags:
Package:         %s
System options:  %s
`,
							aurora.Bold(configurator.Package.Package),
							aurora.Bold(strings.Join(configurator.Package.GetAllOptions("--"), " ")),
						))
				} else {
					f = filepath.Join(homeDir, configurator.Engine.Name+".yml")

					log.InfoC(
						fmt.Sprintf(`Configuring browser start flags:
Package:         %s
User options:    %s
`,
							aurora.Bold(configurator.Package.Package),
							aurora.Bold(strings.Join(configurator.Package.GetAllOptions("--"), " ")),
						))
				}

				// TODO: Add the possibility to remove a user or system yaml if it's
				//       without options.

				if !onlyUpdateIncludes {
					err = configurator.Engine.WriteConfig(f)
					if err != nil {
						log.Fatal("Error on write engine file:", err.Error())
					}

					log.InfoC(fmt.Sprintf("Generated engine file:\t%s", f))
				}

				if binary {

					if system {

						// Write system include
						systemInclude := filepath.Join(systemDir,
							fmt.Sprintf("%s.%s.inc", configurator.Engine.Name,
								filepath.Base(configurator.Package.Binary)))

						err = configurator.GenerateIncludeScript(systemInclude)
						if err != nil {
							log.Fatal(fmt.Sprintf(
								"Error on generate include script %s:", systemInclude,
								err.Error()))
						}

						log.InfoC(fmt.Sprintf("Generated include file:\t%s", systemInclude))

						err = configurator.GenerateScript(opts)
						if err != nil {
							log.Fatal(
								"Error on generate binary script:", err.Error())
						}

						log.InfoC(fmt.Sprintf("Generated script file:\t%s",
							configurator.Package.Binary))
					} else {

						// Write user include
						homeInclude := filepath.Join(homeDir,
							fmt.Sprintf("%s.%s.inc", configurator.Engine.Name,
								filepath.Base(configurator.Package.Binary)))

						err = configurator.GenerateIncludeScript(homeInclude)
						if err != nil {
							log.Fatal(fmt.Sprintf(
								"Error on generate include script %s:", homeInclude,
								err.Error()))
						}

						log.InfoC(fmt.Sprintf("Generated include file:\t%s", homeInclude))
					}

				}

			}

			log.Info("All done.")
		},
	}

	// Ignoring errors to avoid exceptions on
	// containers running macaronictl env-update.
	// Without homedir the command will be interrupted
	// on PreRun.
	homeDir, _ := os.UserHomeDir()

	browserConfigsHomedir := ""
	if homeDir != "" {
		browserConfigsHomedir = filepath.Join(homeDir, ".local/share/macaroni/browsers")
	}

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
	flags.Bool("only-update-includes", false, "Update script includes file.")
	flags.Bool("defaults", false, "Set catalog defaults options to specified package.")
	flags.Bool("without-opts", false, "Disable all options to specified package.")
	flags.Bool("purge", false, "Remove system option from system. Need root permissions.")

	return c
}
