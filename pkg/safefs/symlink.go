package safefs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/errors"
)

// CheckSymlinks rejects symlinks escaping the source directory.
func CheckSymlinks(root, srcDirRel, srcSkill string) error {
	allowed, allowedResolved := resolveAllowed(root, srcDirRel)

	return filepath.WalkDir(srcSkill, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Type()&fs.ModeSymlink == 0 {
			return nil
		}

		target, ok := resolveTarget(path)
		if !ok {
			return nil
		}

		if !isWithin(target, allowed, allowedResolved) {
			rel, _ := filepath.Rel(root, path)

			return errors.WithStack(fmt.Errorf("symlink escapes %s: %s", srcDirRel, filepath.ToSlash(rel)))
		}

		return nil
	})
}

func resolveAllowed(root, srcDirRel string) (allowed, allowedResolved string) {
	allowed = filepath.Join(root, filepath.FromSlash(srcDirRel))

	allowedResolved = allowed
	if resolved, err := filepath.EvalSymlinks(allowed); err == nil {
		allowedResolved = resolved
	}

	return allowed, allowedResolved
}

func resolveTarget(path string) (string, bool) {
	if target, err := filepath.EvalSymlinks(path); err == nil {
		return target, true
	}

	const maxHops = 40
	seen := make(map[string]bool, maxHops)
	cur := path

	for range maxHops {
		if seen[cur] {
			return "", false
		}
		seen[cur] = true

		fi, err := os.Lstat(cur)
		if err != nil {
			return filepath.Clean(cur), true
		}
		if fi.Mode()&fs.ModeSymlink == 0 {
			if target, err := filepath.EvalSymlinks(cur); err == nil {
				return target, true
			}

			return filepath.Clean(cur), true
		}

		link, err := os.Readlink(cur)
		if err != nil {
			return "", false
		}
		if !filepath.IsAbs(link) {
			link = filepath.Join(filepath.Dir(cur), link)
		}

		cur = filepath.Clean(link)
	}

	return "", false
}

func isWithin(target, allowed, allowedResolved string) bool {
	if target == allowed || target == allowedResolved {
		return true
	}

	if strings.HasPrefix(target, allowed+string(os.PathSeparator)) {
		return true
	}

	return strings.HasPrefix(target, allowedResolved+string(os.PathSeparator))
}
