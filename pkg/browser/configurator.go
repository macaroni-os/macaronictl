/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package browser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

const (
	shWrapper = `#!/bin/bash
# Copyright © 2021-2023 Macaroni OS Linux
# Description: Wrapper for package %s generated with macaronictl.

if [ -n "$DEBUG" ] ; then
  set -x
fi

source="%s"
binname="%s"
engine="%s"
homedir="${HOME}/%s"
systemdir="%s"
opts=""

# Check if exists an user configuration
if [ -f "${homedir}/${engine}.${binname}.inc" ] ; then
  source ${homedir}/${engine}.${binname}.inc
else
# Check if exists a system configuration
	if [ -f "${systemdir}/${engine}.${binname}.inc" ] ; then
		source ${systemdir}/${engine}.${binname}.inc
	fi
fi

${source} ${opts}
exit $?
`

	shInclude = `# Copyright © 2021-2023 Macaroni OS Linux
# Description: Startup options of the package %s generated with macaronictl.

opts="%s"
`
)

type BrowserConfiguratorOpts struct {
	Systemdir   string
	Homedir     string
	Catalogfile string
}

type BrowserConfigurator struct {
	Package *BrowserPackage
	Engine  *BrowserEngine
	Catalog *BrowsersCatalog

	SystemConfig *BrowsersCatalog
	HomeConfig   *BrowsersCatalog
}

func NewBrowserConfigurator(opts *BrowserConfiguratorOpts) (*BrowserConfigurator, error) {
	var err error
	ans := &BrowserConfigurator{}

	// Read options catalog
	ans.Catalog, err = LoadBrowsersCatalog(opts.Catalogfile)
	if err != nil {
		return ans, err
	}

	// Load system config
	ans.SystemConfig, _, err = ReadBrowsersOptions(opts.Systemdir)
	if err != nil {
		return ans, err
	}

	// Load user config
	ans.HomeConfig, _, err = ReadBrowsersOptions(opts.Homedir)
	if err != nil {
		return ans, err
	}

	return ans, nil
}

func (c *BrowserConfigurator) GetSystemConfig() *BrowsersCatalog { return c.SystemConfig }
func (c *BrowserConfigurator) GetHomeConfig() *BrowsersCatalog   { return c.HomeConfig }
func (c *BrowserConfigurator) GetCatalog() *BrowsersCatalog      { return c.Catalog }

func (c *BrowserConfigurator) GenerateIncludeScript(fpath string) error {
	if c.Package == nil || c.Engine == nil {
		return errors.New("No package or engine configured to generate scripts.")
	}

	// Add double dash to every options.
	// NOTE: FWIS all options are with two dash.
	opts := c.Package.GetAllOptions("--")
	optsEnv := strings.Join(opts, " ")

	content := fmt.Sprintf(shInclude, c.Package.Package, optsEnv)

	// Remove link
	if utils.Exists(fpath) {
		err := os.Remove(fpath)
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(fpath, []byte(content), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (c *BrowserConfigurator) GenerateScript(opts *BrowserConfiguratorOpts) error {
	if c.Package == nil || c.Engine == nil {
		return errors.New("No package or engine configured to generate scripts.")
	}

	// Remove link
	if utils.Exists(c.Package.Binary) {
		err := os.Remove(c.Package.Binary)
		if err != nil {
			return err
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sanitizedHomedir := opts.Homedir

	if homeDir != "" && strings.HasPrefix(opts.Homedir, homeDir) {
		sanitizedHomedir = opts.Homedir[len(homeDir)+1:]
	}

	// Create the script
	content := fmt.Sprintf(shWrapper,
		c.Package.Package,
		c.Package.Source,
		filepath.Base(c.Package.Binary),
		c.Engine.Name,
		sanitizedHomedir,
		opts.Systemdir,
	)

	err = os.WriteFile(c.Package.Binary, []byte(content), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (c *BrowserConfigurator) WorkingOnPackage(pkgname string, system bool) error {
	engineCat, pkgCat := c.Catalog.GetEngineAndPackage(pkgname)
	if engineCat == nil {
		return fmt.Errorf(
			"Package %s not found on catalog %s",
			pkgname)
	}

	if system {
		// POST: Working on package to setup system options

		engine, pkg := c.SystemConfig.GetEngineAndPackage(pkgname)
		if pkg == nil {
			// POST: No system configurations availables or home configurations.
			//       Using the catalog defaults.
			c.Package = pkgCat
			if engine != nil {
				c.Engine = engine
			} else {
				c.Engine = NewBrowserEngine(engineCat.Name)
				c.Engine.SetPackage(pkgCat)
			}
		} else {
			c.Package = pkg
			c.Engine = engine
		}

	} else {
		// POST: Working on package to setup user options

		engine, pkg := c.HomeConfig.GetEngineAndPackage(pkgname)
		if pkg == nil {
			// POST: No system configurations availables or home configurations.
			//       Using the catalog defaults.
			c.Package = pkgCat
			if engine != nil {
				c.Engine = engine
			} else {
				c.Engine = NewBrowserEngine(engineCat.Name)
				c.Engine.SetPackage(pkgCat)
			}
		} else {
			c.Package = pkg
			c.Engine = engine
		}

	}

	return nil
}
