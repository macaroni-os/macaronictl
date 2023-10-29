/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package browser

import (
	"os"
	"path/filepath"

	"github.com/macaroni-os/macaronictl/pkg/utils"

	"gopkg.in/yaml.v3"
)

func NewBrowsersCatalog() *BrowsersCatalog {
	return &BrowsersCatalog{
		Engines: make(map[string]*BrowserEngine, 0),
	}
}

func LoadBrowsersCatalog(fname string) (*BrowsersCatalog, error) {
	ans := NewBrowsersCatalog()
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	return ans, nil
}

func (c *BrowsersCatalog) GetPackage(pname string) *BrowserPackage {
	for _, be := range c.Engines {
		if bp, ok := be.GetPackage(pname); ok {
			return bp
		}
	}
	return nil
}

func (c *BrowsersCatalog) IsEmpty() bool {
	return len(c.Engines) == 0
}

func (c *BrowsersCatalog) GetEngine(ename string) *BrowserEngine {
	for e, be := range c.Engines {
		if e == ename {
			return be
		}
	}
	return nil
}

func (c *BrowsersCatalog) GetEngineAndPackage(pname string) (
	*BrowserEngine, *BrowserPackage) {

	for i, _ := range c.Engines {
		p, present := c.Engines[i].GetPackage(pname)
		if present {
			return c.Engines[i], p
		}
	}

	return nil, nil
}

func (c *BrowsersCatalog) Merge(o *BrowsersCatalog) {
	for e, be := range o.Engines {
		cbe := c.GetEngine(e)
		if cbe != nil {
			cbe.Merge(be)
		} else {
			c.Engines[e] = be
		}
	}
}

func (c *BrowsersCatalog) MergeEngine(e *BrowserEngine) {
	cbe := c.GetEngine(e.Name)
	if cbe != nil {
		cbe.Merge(e)
	} else {
		c.Engines[e.Name] = e
	}
}

func (c *BrowsersCatalog) WriteEngineConfigs(dirname string) error {
	if !utils.Exists(dirname) {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	for k, _ := range c.Engines {
		engineFile := filepath.Join(dirname, k+".yml")
		err := c.Engines[k].WriteConfig(engineFile)
		if err != nil {
			return err
		}
	}

	return nil
}
