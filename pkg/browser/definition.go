/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package browser

type BrowserEngine struct {
	Name     string                     `yaml:"engine" json:"engine"`
	Packages map[string]*BrowserPackage `yaml:"packages,omitempty" json:"packages,omitempty"`
	Options  []*BrowserOpt              `yaml:"options,omitempty" json:"options,omitempty"`
}

type BrowserPackage struct {
	Package string `yaml:"package" json:"package"`
	Binary  string `yaml:"binary" json:"binary"`
	Source  string `yaml:"source" json:"source"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`

	EnabledOptions []*BrowserOpt `yaml:"enabled_opts,omitempty" json:"enabled_opts,omitempty"`
}

type BrowserOpt struct {
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Default     bool     `yaml:"default,omitempty" json:"default,omitempty"`
	Option      []string `yaml:"option" json:"option"`
}

type BrowsersCatalog struct {
	Engines map[string]*BrowserEngine `yaml:"engines,omitempty" json:"engines,omitempty"`
}
