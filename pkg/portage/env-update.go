/*
Copyright © 2021-2022 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package portage

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/logger"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

type EnvUpdateOpts struct {
	EnvSpaceSeparated []string
	EnvColonSeparated []string
	EnvSingleValue    []string
	WithLdConfig      bool
	PrelinkCapable    bool
	DryRun            bool
	Debug             bool
	Systemd           bool
	Csh               bool
}

const (
	profileEnvFileHeader = `# THIS FILE IS AUTOMATICALLY GENERATED BY macaronictl env-update.
# DO NOT EDIT THIS FILE. CHANGES TO STARTUP PROFILES
# GO INTO /etc/profile NOT /etc/profile.env
`
	ldsoConfHeader = `# ld.so.conf autogenerated by macaronictl env-update; make all changes to
# contents of /etc/env.d directory
`
)

var (
	envSkipped = []string{
		"COLON_SEPARATED",
		"LDPATH",
	}
)

func NewEnvUpdateOpts() *EnvUpdateOpts {
	return &EnvUpdateOpts{
		EnvSpaceSeparated: []string{
			"CONFIG_PROTECT",
			"CONFIG_PROTECT_MASK",
		},
		EnvColonSeparated: []string{
			"ADA_INCLUDE_PATH",
			"ADA_OBJECTS_PATH",
			"CLASSPATH",
			"INFODIR",
			"INFOPATH",
			"KDEDIRS",
			"LDPATH",
			"MANPATH",
			"PATH",
			"PKG_CONFIG_PATH",
			"PRELINK_PATH",
			"PRELINK_PATH_MASK",
			"PYTHONPATH",
			"ROOTPATH",
			// Set static variables instead of read COLON_SEPARATED variable.
			"XDG_DATA_DIRS",
			"XDG_CONFIG_DIRS",
		},
		EnvSingleValue: []string{
			"LANG",
			"GSETTINGS_BACKEND",
		},
		WithLdConfig:   true,
		PrelinkCapable: true,
		DryRun:         false,
		Debug:          false,
		Systemd:        false,
		Csh:            false,
	}
}

func CheckLocale(mRef *map[string]string) error {
	// TODO
	return nil
}

func writeProfileEnv(file string, opts *EnvUpdateOpts,
	mRef *map[string]string) error {
	var f *os.File
	var err error

	envs := *mRef

	// Sort envs keys
	keys := []string{}
	for k := range envs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if opts.DryRun {
		// On dry run I just print the variables to stdout.
		f = os.Stdout
	} else {
		f, err = os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
	}

	_, err = f.WriteString(profileEnvFileHeader + "\n")
	if err != nil {
		return err
	}

	for _, k := range keys {
		v := envs[k]
		if strings.HasPrefix(v, "$") && (!strings.HasPrefix(v, "${")) {
			_, err = f.WriteString(fmt.Sprintf("export %s=$'%s'\n", k, envs[k]))
		} else {
			_, err = f.WriteString(fmt.Sprintf("export %s='%s'\n", k, envs[k]))
		}
		if err != nil {
			return err
		}
	}

	if !opts.DryRun {
		defer f.Close()
	}

	return nil
}

func writeLdsoConf(file, ldpath string, opts *EnvUpdateOpts) error {
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

	_, err = f.WriteString(ldsoConfHeader + "\n")
	if err != nil {
		return err
	}

	atoms := strings.Split(ldpath, ":")
	for _, k := range atoms {
		_, err = f.WriteString(k + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func execLdconfig(rootdir, ldpath string, opts *EnvUpdateOpts) error {
	log := logger.GetDefaultLogger()
	// Retrieve the abs path of ldconfig to avoid errors
	// when the $PATH is not set correctly.
	ldconfig := utils.TryResolveBinaryAbsPath("ldconfig")

	if ldconfig == "ldconfig" {
		log.Warning("ldconfig on defined paths. Try to run it without abs path.")
	} else {
		log.Debug("Using ldconfig path:", ldconfig)
	}

	args := []string{
		"-X", "-r", fmt.Sprintf("%s", rootdir),
	}
	if opts.Debug {
		args = append(args, "-v")
	}

	// TODO: check if create ldconfig binary from CHOST
	ldconfigCommand := exec.Command(ldconfig, args...)
	ldconfigCommand.Stdout = os.Stdout
	ldconfigCommand.Stderr = os.Stderr
	ldconfigCommand.Env = []string{
		fmt.Sprintf("LDPATH=%s", ldpath),
	}

	err := ldconfigCommand.Start()
	if err != nil {
		return errors.New("Error on start ldconfig command: " + err.Error())
	}

	err = ldconfigCommand.Wait()
	if err != nil {
		return errors.New("Error on waiting ldconfig command: " + err.Error())
	}

	if ldconfigCommand.ProcessState.ExitCode() != 0 {
		return fmt.Errorf(
			"ldconfig command exiting with %d",
			ldconfigCommand.ProcessState.ExitCode())
	}

	return nil

}

func EnvUpdate(rootdir string, opts *EnvUpdateOpts) error {
	log := logger.GetDefaultLogger()

	// Parse env.d directory
	envs, err := ParseEnvd(rootdir, opts)
	if err != nil {
		return err
	}

	// Check locale
	err = CheckLocale(&envs)
	if err != nil {
		return err
	}

	// retrieve LD_PATH for ldconfig execution.
	ldpath := envs["LDPATH"]

	// Exclude skipped envs
	sanitizedEnvs := make(map[string]string, 0)
	for k, v := range envs {
		skipValue := utils.KeyInList(k, &envSkipped)
		if skipValue {
			continue
		}
		sanitizedEnvs[k] = v
	}

	profileEnvFile := filepath.Join(rootdir, "/etc/profile.env")
	log.Info(fmt.Sprintf(">>> Generating %s...", profileEnvFile))
	// Write or print file /etc/profile.env
	err = writeProfileEnv(profileEnvFile, opts, &envs)
	if err != nil {
		return err
	}

	if opts.PrelinkCapable {
		prelinkConfDir := filepath.Join(rootdir, "/etc/prelink.conf.d")
		if utils.Exists(prelinkConfDir) {
			prelinkConfFile := filepath.Join(prelinkConfDir, "/portage.conf")
			log.Info(fmt.Sprintf(">>> Generating %s...", prelinkConfFile))

			ppath, pmpaths := preparePrelinkPaths(&envs, opts)
			// Write prelink.conf.d/portage.conf
			err = writePrelinkFile(prelinkConfFile, ppath, pmpaths, opts)
			if err != nil {
				return err
			}
		}
	}

	if opts.Systemd {
		systemdEnvFile := filepath.Join(rootdir,
			"/etc/environment.d/10-macaroni-env.conf")
		log.Info(fmt.Sprintf(">>> Generating %s...", systemdEnvFile))

		// Write systemd env file
		err = writeSystemdEnvFile(systemdEnvFile, &envs, opts)
		if err != nil {
			return err
		}
	}

	if opts.WithLdConfig {
		if ldpath == "" {
			log.Warning("Find empty LDPATH. Using default values.")
			// Warning no ldpath set.
			ldpath = "include ld.so.conf.d/*.conf:/lib:/usr/lib:/usr/local/lib"
		}

		ldsoconfFile := filepath.Join(rootdir, "/etc/ld.so.conf")
		log.Info(fmt.Sprintf(">>> Generating %s...", ldsoconfFile))
		err := writeLdsoConf(ldsoconfFile, ldpath, opts)
		if err != nil {
			return err
		}

		if !opts.DryRun {

			log.Info(fmt.Sprintf(
				">>> Regenerating %s...", filepath.Join(rootdir, "/etc/ld.so.cache")))
			err := execLdconfig(rootdir, ldpath, opts)
			if err != nil {
				return err
			}

		}

	}

	if opts.Csh {
		cshEnvfile := filepath.Join(rootdir, "/etc/csh.env")
		log.Info(fmt.Sprintf(">>> Generating %s...", cshEnvfile))

		// Write /etc/csh.env file
		err = writeCshEnvFile(cshEnvfile, &envs, opts)
		if err != nil {
			return err
		}
	}

	return nil
}
