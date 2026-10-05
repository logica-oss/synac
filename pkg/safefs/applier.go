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

// Applier writes filesystem changes, honoring dry-run.
type Applier struct {
	log    *slog.Logger
	root   *os.Root
	dryRun bool
}

// NewApplier creates an Applier bound to root.
func NewApplier(log *slog.Logger, root string, dryRun bool) (*Applier, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("open root: %w", err))
	}

	return &Applier{log: log, root: r, dryRun: dryRun}, nil
}

// Close releases the root handle.
func (a *Applier) Close() error {
	return a.root.Close()
}

// Within rejects paths escaping the root.
func (a *Applier) Within(destRel string) error {
	if _, err := a.root.Lstat(destRel); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return errors.WithStack(fmt.Errorf("validate dest %s: %w", destRel, err))
	}

	return nil
}

// WriteFile writes data to dest, creating parents as needed.
func (a *Applier) WriteFile(dest, destRel string, data []byte, perm fs.FileMode) error {
	if err := a.Within(destRel); err != nil {
		return err
	}

	if a.dryRun {
		a.log.Info("dry-run: would write", "path", destRel)
		return nil
	}

	if dir := filepath.Dir(destRel); dir != "." {
		if err := a.root.MkdirAll(dir, 0o755); err != nil {
			return errors.WithStack(fmt.Errorf("create parent dir: %w", err))
		}
	}

	if err := a.root.WriteFile(destRel, data, perm); err != nil {
		return errors.WithStack(fmt.Errorf("write %s: %w", destRel, err))
	}

	return nil
}

// MkdirAll creates dest and parents.
func (a *Applier) MkdirAll(dest, destRel string, perm fs.FileMode) error {
	if err := a.Within(destRel); err != nil {
		return err
	}

	if a.dryRun {
		a.log.Info("dry-run: would create dir", "path", destRel)
		return nil
	}

	if err := a.root.MkdirAll(destRel, perm); err != nil {
		return errors.WithStack(fmt.Errorf("create dest dir %s: %w", destRel, err))
	}

	return nil
}

// RemoveAll removes dest.
func (a *Applier) RemoveAll(dest, destRel string) error {
	if err := a.Within(destRel); err != nil {
		return err
	}

	if a.dryRun {
		a.log.Info("dry-run: would remove", "path", destRel)
		return nil
	}

	if err := a.root.RemoveAll(destRel); err != nil {
		return errors.WithStack(fmt.Errorf("clear dest dir %s: %w", destRel, err))
	}

	return nil
}

// CopyDir copies src to dest, keeping symlinks as links.
func (a *Applier) CopyDir(src, dest, destRel string) error {
	if err := a.Within(destRel); err != nil {
		return err
	}

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
