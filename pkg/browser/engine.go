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

func NewBrowserEngine(name string) *BrowserEngine {
	return &BrowserEngine{
		Name:     name,
		Packages: make(map[string]*BrowserPackage, 0),
		Options:  []*BrowserOpt{},
	}
}

func (e *BrowserEngine) Merge(m *BrowserEngine) {
	// Fornow just merge packages options
	for k := range m.Packages {
		if _, ok := e.Packages[k]; ok {
			e.Packages[k].Merge(m.Packages[k])
		} else {
			e.Packages[k] = m.Packages[k]
		}
	}
}

func (e *BrowserEngine) SetPackage(pkg *BrowserPackage) {
	e.Packages[pkg.Package] = pkg
}

func (e *BrowserEngine) NumPackages() int {
	return len(e.Packages)
}

func (e *BrowserEngine) GetPackage(pname string) (*BrowserPackage, bool) {
	ans, ok := e.Packages[pname]
	return ans, ok
}

func (e *BrowserEngine) GetDefaultOptions() []*BrowserOpt {
	ans := []*BrowserOpt{}
	for i := range e.Options {
		if e.Options[i].Default {
			ans = append(ans, e.Options[i])
		}
	}
	return ans
}

func (e *BrowserEngine) WriteConfig(f string) error {
	optsTmp := e.Options
	e.Options = []*BrowserOpt{}

	dirname := filepath.Dir(f)
	if !utils.Exists(dirname) {
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		e.Options = optsTmp
		return err
	}

	err = os.WriteFile(f, data, 0644)
	e.Options = optsTmp

	return err
}
