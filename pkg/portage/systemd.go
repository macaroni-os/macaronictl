/*
Copyright © 2021-2022 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package portage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/macaroni-os/macaronictl/pkg/utils"
)

const (
	systemdEnvFileHeader = `# THIS FILE IS AUTOMATICALLY GENERATED BY macaronictl env-update.
# DO NOT EDIT THIS FILE. CHANGES TO STARTUP PROFILES
# GO INTO /etc/profile NOT /etc/profile.env

`
)

func writeSystemdEnvFile(file string, mRef *map[string]string, opts *EnvUpdateOpts) error {
	var f *os.File
	var err error

	envdir := filepath.Dir(file)
	if !utils.Exists(envdir) {
		err := os.MkdirAll(envdir, 0750)
		if err != nil {
			return err
		}
	}

	if opts.DryRun {
		f = os.Stdout
	} else {
		f, err = os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	// Write file header
	_, err = f.WriteString(systemdEnvFileHeader + "\n")
	if err != nil {
		return err
	}

	envs := *mRef

	// Sort envs keys
	keys := []string{}
	for k := range envs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if k == "LDPATH" {
			continue
		}

		// Systemd doesn't accept empty values
		if envs[k] == "" {
			continue
		}

		_, err = f.WriteString(fmt.Sprintf("%s=%s\n", k, envs[k]))
		if err != nil {
			return err
		}
	}

	return nil
}
