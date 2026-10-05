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

	link, err := os.Readlink(path)
	if err != nil {
		return "", false
	}

	if !filepath.IsAbs(link) {
		dir := filepath.Dir(path)
		if resolvedDir, rErr := filepath.EvalSymlinks(dir); rErr == nil {
			dir = resolvedDir
		}

		link = filepath.Join(dir, link)
	}

	return filepath.Clean(link), true
}

// CheckWithinRoot rejects paths escaping root through symlinks.
func CheckWithinRoot(root, path, pathRel string) error {
	rootResolved := root
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		rootResolved = resolved
	}

	target := path
	var rest []string
	for {
		if _, err := os.Lstat(target); err == nil {
			break
		}

		parent := filepath.Dir(target)
		if parent == target {
			return errors.WithStack(fmt.Errorf("resolve path %s: no existing ancestor", filepath.ToSlash(pathRel)))
		}
		rest = append([]string{filepath.Base(target)}, rest...)
		target = parent
	}

	resolved, ok := resolveTarget(target)
	if !ok {
		return errors.WithStack(fmt.Errorf("resolve path %s: unable to resolve symlink", filepath.ToSlash(pathRel)))
	}

	full := resolved
	for _, p := range rest {
		full = filepath.Join(full, p)
	}

	if !isWithin(filepath.Clean(full), root, rootResolved) {
		return errors.WithStack(fmt.Errorf("path escapes repository root: %s", filepath.ToSlash(pathRel)))
	}

	return nil
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
