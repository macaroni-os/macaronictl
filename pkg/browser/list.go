/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/macaroni-os/macaronictl/pkg/anise"
	"github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
	"gopkg.in/yaml.v3"
)

func AvailableBrowsers(installed bool,
	config *specs.MacaroniCtlConfig) (*specs.StonesPack, error) {

	luet := utils.TryResolveBinaryAbsPath("luet")
	args := []string{
		luet, "search", "-a", "desktop_browser",
		"-o", "json",
	}

	if installed {
		args = append(args, "--installed")
	}

	return anise.SearchStones(args)
}

func ReadBrowserEngine(fname string) (*BrowserEngine, error) {
	ans := NewBrowserEngine("")
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	return ans, nil
}

func ReadBrowsersOptions(dirname string) (*BrowsersCatalog, bool, error) {
	var regexConf = regexp.MustCompile(`.yml$|.yaml$`)
	configPresent := false
	ans := NewBrowsersCatalog()

	if !utils.Exists(dirname) {
		return ans, configPresent, nil
	}

	isDir, err := utils.IsDir(dirname)
	if err != nil {
		return ans, configPresent, fmt.Errorf(
			"Error on check if %s is a directory: %s",
			dirname, err.Error())
	} else if !isDir {
		return ans, configPresent, fmt.Errorf(
			"Browser options directory %s used is not a directory",
			dirname)
	}

	dirEntries, err := os.ReadDir(dirname)
	if err != nil {
		return ans, configPresent, err
	}

	for _, file := range dirEntries {
		if file.IsDir() {
			continue
		}

		if !regexConf.MatchString(file.Name()) {
			continue
		}

		loadedEngine, err := ReadBrowserEngine(
			filepath.Join(dirname, file.Name()))
		if err != nil {
			return ans, configPresent, err
		}

		ans.MergeEngine(loadedEngine)
	}

	return ans, configPresent, nil
}
