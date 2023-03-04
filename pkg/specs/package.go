/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

type Stone struct {
	Name        string                 `json:"name" yaml:"name"`
	Category    string                 `json:"category" yaml:"category"`
	Version     string                 `json:"version" yaml:"version"`
	License     string                 `json:"license,omitempty" yaml:"license,omitempty"`
	Repository  string                 `json:"repository,omitempty" yaml:"repository"`
	Hidden      bool                   `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	Files       []string               `json:"files,omitempty" yaml:"files,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	UseFlags    []string               `json:"use_flags,omitempty" yaml:"use_flags,omitempty"`

	Provides  []*Stone `json:"provides,omitempty" yaml:"provides,omitempty"`
	Requires  []*Stone `json:"requires,omitempty" yaml:"requires,omitempty"`
	Conflicts []*Stone `json:"conflicts,omitempty" yaml:"conflicts,omitempty"`
}

type StonesPack struct {
	Stones []*Stone `json:"stones" yaml:"stones"`
}

type KernelAnnotation struct {
	EoL      string `json:"eol,omitempty" yaml:"eol,omitempty"`
	Lts      bool   `json:"lts" yaml:"lts"`
	Released string `json:"released,omitempty" yaml:"released,omitempty"`
	Suffix   string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	Type     string `json:"vanilla,omitempty" yaml:"vanilla,omitempty"`
}