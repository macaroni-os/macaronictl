/*
Copyright © 2021-2022 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package utils

import (
	"os"
)

func KeyInList(key string, arr *[]string) bool {
	for _, k := range *arr {
		if k == key {
			return true
		}
	}

	return false
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
