// Copyright © 2021 Daniele Rondina, geaaru@sabayonlinux.org
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

func OsRelease() (string, error) {
	mosReleaseFile := "/etc/mocaccino/release"
	release := ""

	_, err := os.Stat(mosReleaseFile)
	if err == nil {
		content, err := ioutil.ReadFile(mosReleaseFile)
		if err != nil {
			return "", err
		}

		release = string(content)
		release = strings.ReplaceAll(release, "\n", "")
	} else if !os.IsNotExist(err) {
		return release, err
	} // else is not a mocaccino os rootfs.

	return release, nil
}
