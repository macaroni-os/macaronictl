/*
Copyright © 2021-2022 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package portage

import (
	"fmt"
	"os"
	"sort"
)

const (
	cshEnvFileHeader = `# THIS FILE IS AUTOMATICALLY GENERATED BY macaronictl env-update.
# DO NOT EDIT THIS FILE. CHANGES TO STARTUP PROFILES
# GO INTO /etc/csh.cshrc NOT /etc/csh.env

`
)

// Create the file /etc/csh.env for (t)csh support
func writeCshEnvFile(file string, mRef *map[string]string, opts *EnvUpdateOpts) error {
	var f *os.File
	var err error

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
	_, err = f.WriteString(cshEnvFileHeader + "\n")
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

		_, err = f.WriteString(fmt.Sprintf("setenv %s '%s'\n", k, envs[k]))
		if err != nil {
			return err
		}
	}

	return nil
}
