package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

var errNoModule = errors.New("no go.mod found")

type module struct {
	dir   string
	goMod string
}

func findModules(root string) ([]module, error) {
	ignored, err := loadGitignore(root)
	if err != nil {
		return nil, err
	}

	var modules []module
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if shouldSkipDir(ignored, root, path, d.Name()) {
				return filepath.SkipDir
			}

			return nil
		}
		if d.Name() != "go.mod" {
			return nil
		}

		dir := filepath.Dir(path)
		modules = append(modules, module{dir: dir, goMod: path})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func loadGitignore(root string) (*ignore.GitIgnore, error) {
	path := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return ignore.CompileIgnoreLines(), nil
		}

		return nil, err
	}

	return ignore.CompileIgnoreFile(path)
}

func shouldSkipDir(ignored *ignore.GitIgnore, root, path, name string) bool {
	if name == ".git" {
		return true
	}

	if ignored == nil {
		return false
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	return ignored.MatchesPath(rel)
}
