// Package safefs applies filesystem changes with dry-run support.
package safefs

import (
	"os"
)

// SameFile reports whether two paths refer to the same file.
func SameFile(first, second string) bool {
	aInfo, aErr := os.Stat(first)
	if aErr != nil {
		return false
	}

	bInfo, bErr := os.Stat(second)
	if bErr != nil {
		return false
	}

	return os.SameFile(aInfo, bInfo)
}
