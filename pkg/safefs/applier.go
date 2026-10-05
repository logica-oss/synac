package safefs

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
	"github.com/otiai10/copy"
)

type Applier struct {
	log    *slog.Logger
	dryRun bool
}

func NewApplier(log *slog.Logger, dryRun bool) *Applier {
	return &Applier{log: log, dryRun: dryRun}
}

func (a *Applier) WriteFile(dest, destRel string, data []byte, perm fs.FileMode) error {
	if a.dryRun {
		a.log.Info("dry-run: would write", "path", destRel)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return errors.WithStack(fmt.Errorf("create parent dir: %w", err))
	}

	if err := os.WriteFile(dest, data, perm); err != nil {
		return errors.WithStack(fmt.Errorf("write %s: %w", destRel, err))
	}

	return nil
}

func (a *Applier) MkdirAll(dest, destRel string, perm fs.FileMode) error {
	if a.dryRun {
		a.log.Info("dry-run: would create dir", "path", destRel)
		return nil
	}

	if err := os.MkdirAll(dest, perm); err != nil {
		return errors.WithStack(fmt.Errorf("create dest dir %s: %w", destRel, err))
	}

	return nil
}

func (a *Applier) RemoveAll(dest, destRel string) error {
	if a.dryRun {
		a.log.Info("dry-run: would remove", "path", destRel)
		return nil
	}

	if err := os.RemoveAll(dest); err != nil {
		return errors.WithStack(fmt.Errorf("clear dest dir %s: %w", destRel, err))
	}

	return nil
}

func (a *Applier) CopyDir(src, dest, destRel string) error {
	if a.dryRun {
		a.log.Info("dry-run: would copy dir", "path", destRel)
		return nil
	}

	return copy.Copy(src, dest, copy.Options{
		OnSymlink: func(string) copy.SymlinkAction {
			return copy.Shallow
		},
	})
}
