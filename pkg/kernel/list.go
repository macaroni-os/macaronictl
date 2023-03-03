/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func AvailableKernels(config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {
	var errBuffer bytes.Buffer
	var outBuffer bytes.Buffer
	var ans specs.StonesPack

	args := []string{
		"luet", "search", "-a", "kernel", "-n", "macaroni-full",
		"-o", "json",
	}

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdout = utils.NewNopCloseWriter(&outBuffer)
	cmd.Stderr = utils.NewNopCloseWriter(&errBuffer)

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return nil, fmt.Errorf("luet search exiting with %s: %s",
			cmd.ProcessState.ExitCode(),
			errBuffer.String())
	}

	// Read json output.
	err = json.Unmarshal(outBuffer.Bytes(), &ans)
	if err != nil {
		return nil, fmt.Errorf("Error on unmarshal json data: %s",
			err.Error())
	}

	return &ans, nil
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
