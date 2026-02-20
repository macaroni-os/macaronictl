/*
Copyright © 2021-2022 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package portage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/logger"
	"github.com/macaroni-os/macaronictl/pkg/utils"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func ParseEnvFile(file string) (map[string]string, error) {
	log := logger.GetDefaultLogger()
	ans := make(map[string]string, 0)

	log.Debug("Parsing env file ", file)

	f, err := os.Open(file)
	if err != nil {
		log.Error("Error on open file", file,
			": ", err.Error())
		return ans, err
	}
	defer f.Close()

	node, err := syntax.NewParser().Parse(f, file)
	if err != nil {
		return ans, fmt.Errorf("Error on parse: %v", err)
	}

	r, _ := interp.New(
		interp.Env(expand.ListEnviron("")),
	)
	if err := r.Run(context.Background(), node); err != nil {
		return ans, fmt.Errorf("Error on run file %s: %v",
			file, err)
	}

	// delete the internal shell vars that the user is not
	// interested in
	delete(r.Vars, "PWD")
	delete(r.Vars, "OLDPWD")
	delete(r.Vars, "PPID")
	delete(r.Vars, "HOME")
	delete(r.Vars, "IFS")
	delete(r.Vars, "OPTIND")
	delete(r.Vars, "GID")
	delete(r.Vars, "UID")
	delete(r.Vars, "EUID")

	for k, v := range r.Vars {
		ans[k] = v.String()
	}

	return ans, nil
}

// Merge env values
func mergeEnvValues(newValue, prevValue, sep string) string {
	vPrev := strings.Split(prevValue, sep)
	vN := strings.Split(newValue, sep)

	m := make(map[string]bool, 0)
	values := []string{}

	// Note: It's possible that a previous value is empty
	// LDPATH=""
	if prevValue == "" {
		return newValue
	}

	for _, k := range vPrev {
		m[k] = true
		values = append(values, k)
	}

	for _, k := range vN {
		if _, ok := m[k]; !ok {
			m[k] = true
			values = append(values, k)
		}
	}

	ans := strings.Join(values, sep)
	return ans
}

// Parse /etc/env.d of the selected rootdir to generate the
// list of env variables used to generate /etc/pofile.env, csh.env,
// ld.so.conf, prelink.conf.
func ParseEnvd(rootdir string, opts *EnvUpdateOpts) (map[string]string, error) {
	var regexEnvfiles = regexp.MustCompile(`^[0-9][0-9].*`)
	ans := make(map[string]string, 0)
	log := logger.GetDefaultLogger()

	envDir := filepath.Join(rootdir, "/etc/env.d")

	log.DebugC(fmt.Sprintf(
		"Parsing %s directory to read all env vars.", envDir))

	files, err := ioutil.ReadDir(envDir)
	if err != nil {
		return ans, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !regexEnvfiles.MatchString(file.Name()) {
			log.Debug("File", file.Name(), "skipped.")
			continue
		}

		envFile := filepath.Join(envDir, file.Name())

		// Parse the environment file
		m, err := ParseEnvFile(envFile)
		if err != nil {
			return ans, err
		}

		// Merge map
		for k, v := range m {

			singleValue := utils.KeyInList(k, &opts.EnvSingleValue)
			if singleValue {
				ans[k] = v
				continue
			}

			if val, ok := ans[k]; ok {

				sep := " "
				colonSep := utils.KeyInList(k, &opts.EnvColonSeparated)
				if colonSep {
					sep = ":"
				}

				ans[k] = mergeEnvValues(v, val, sep)
			} else {
				ans[k] = v
			}
		}
	}

	return ans, nil
}
