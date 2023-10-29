/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package browser

import (
	"fmt"
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
