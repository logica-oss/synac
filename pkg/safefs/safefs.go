// Package safefs applies filesystem changes with dry-run support.
package safefs

import (
	"os"
)

// SameFile reports whether two paths refer to the same file.
func SameFile(a, b string) bool {
	aInfo, aErr := os.Stat(a)
	if aErr != nil {
		return false
	}

	bInfo, bErr := os.Stat(b)
	if bErr != nil {
		return false
	}

	return os.SameFile(aInfo, bInfo)
}
