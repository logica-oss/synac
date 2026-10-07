package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/bobg/mingo"
	"github.com/k1LoW/errors"
	"golang.org/x/mod/modfile"
)

func pinGoDirective(mod module) error {
	slog.Info("mingo", "dir", mod.dir)

	scanner := mingo.Scanner{Deps: true, Indirect: true}
	result, err := scanner.ScanDir(mod.dir)
	if err != nil {
		return errors.WithStack(fmt.Errorf("mingo: %w", err))
	}

	declared, err := readGoVersion(mod.goMod)
	if err != nil {
		return err
	}

	minimum := "1." + strconv.Itoa(result.Version()) + ".0"
	if declared == minimum {
		slog.Info("go directive is already minimal", "version", declared)

		return nil
	}

	slog.Info("pin go directive", "from", declared, "to", minimum)

	return writeGoVersion(mod, minimum)
}

func readGoVersion(path string) (string, error) {
	file, err := parseModFile(path)
	if err != nil {
		return "", err
	}
	if file.Go == nil {
		return "", nil
	}

	return file.Go.Version, nil
}

func writeGoVersion(mod module, version string) error {
	file, err := parseModFile(mod.goMod)
	if err != nil {
		return err
	}

	if file.Go != nil {
		file.DropGoStmt()
	}
	if addErr := file.AddGoStmt(version); addErr != nil {
		return errors.WithStack(fmt.Errorf("set go %s: %w", version, addErr))
	}

	file.Cleanup()

	data := modfile.Format(file.Syntax)
	info, err := os.Stat(mod.goMod)
	if err != nil {
		return errors.WithStack(fmt.Errorf("stat go.mod: %w", err))
	}

	if err := os.WriteFile(mod.goMod, data, info.Mode().Perm()); err != nil {
		return errors.WithStack(fmt.Errorf("write go.mod: %w", err))
	}

	return nil
}

func parseModFile(path string) (*modfile.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("read go.mod: %w", err))
	}

	file, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse go.mod: %w", err))
	}

	return file, nil
}
