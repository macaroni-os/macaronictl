/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernel

import (
	"fmt"

	"github.com/macaroni-os/macaronictl/pkg/anise"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func AvailableExtraModules(kernelBranch string, installed bool,
	config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {

	luet := utils.TryResolveBinaryAbsPath("luet")
	args := []string{
		luet, "search", "-a", "kernel_module",
		"-o", "json",
	}

	if kernelBranch != "" {
		args = append(args, []string{
			"--category", "kernel-" + kernelBranch,
		}...)
	}

	if installed {
		args = append(args, "--installed")
	}

	return anise.SearchStones(args)
}

func AvailableKernels(config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {
	luet := utils.TryResolveBinaryAbsPath("luet")
	args := []string{
		luet, "search", "-a", "kernel", "-n", "macaroni-full",
		"-o", "json",
	}

	return anise.SearchStones(args)
}

func InstalledKernels(config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {
	luet := utils.TryResolveBinaryAbsPath("luet")
	args := []string{
		luet, "search", "-a", "kernel", "-n", "macaroni-full",
		"-o", "json", "--installed",
	}

	return anise.SearchStones(args)
}

func ParseKernelAnnotations(s *specs.Stone) (*specs.KernelAnnotation, error) {
	ans := &specs.KernelAnnotation{
		EoL:      "",
		Lts:      false,
		Released: "",
		Suffix:   "",
		Type:     "",
	}

	fieldsI, ok := s.Annotations["kernel"]
	if !ok {
		return nil, fmt.Errorf("[%s/%s] No kernel annotation key found",
			s.Category, s.Name)
	}

	fields, ok := fieldsI.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("[%s/%s] Error on cast annotations fields",
			s.Category, s.Name)
	}

	// Get eol
	if val, ok := fields["eol"]; ok {
		ans.EoL, _ = val.(string)
	}

	// Get lts
	if val, ok := fields["lts"]; ok {
		ans.Lts, _ = val.(bool)
	}

	// Get released
	if val, ok := fields["released"]; ok {
		ans.Released, _ = val.(string)
	}

	// Get suffix
	if val, ok := fields["suffix"]; ok {
		ans.Suffix, _ = val.(string)
	}

	// Get type
	if val, ok := fields["type"]; ok {
		ans.Type, _ = val.(string)
	}

	return ans, nil
}
