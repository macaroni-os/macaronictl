/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package kernel

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/logger"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func InstallPackages(k *specs.Stone, modules []*specs.Stone) error {
	log := logger.GetDefaultLogger()
	luet := utils.TryResolveBinaryAbsPath("luet")
	args := []string{
		luet, "i", k.GetName(),
	}
	for _, s := range modules {
		args = append(args, s.GetName())
	}

	cmd := exec.Command(args[0], args[1:]...)
	log.Debug(fmt.Sprintf("Running install commmand: %s",
		strings.Join(args, " ")))

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("luet install exiting with %s.",
			cmd.ProcessState.ExitCode())
	}

	return nil
}
