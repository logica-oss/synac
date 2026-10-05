package safefs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/errors"
)

func CheckSymlinks(root, srcDirRel, srcSkill string) error {
	allowed := filepath.Join(root, filepath.FromSlash(srcDirRel))

	return filepath.WalkDir(srcSkill, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Type()&fs.ModeSymlink == 0 {
			return nil
		}

		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			link, lErr := os.Readlink(path)
			if lErr != nil {
				return nil
			}

			if !filepath.IsAbs(link) {
				link = filepath.Join(filepath.Dir(path), link)
			}

			target = filepath.Clean(link)
		}

		if target != allowed && !strings.HasPrefix(target, allowed+string(os.PathSeparator)) {
			rel, _ := filepath.Rel(root, path)
			return errors.WithStack(fmt.Errorf("symlink escapes %s: %s", srcDirRel, filepath.ToSlash(rel)))
		}

		return nil
	})
}
