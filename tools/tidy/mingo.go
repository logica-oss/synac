package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bobg/mingo"
	"github.com/k1LoW/errors"
	"golang.org/x/mod/modfile"
)

// pinGoDirective rewrites the go directive to the version mingo computes,
// leaving the file untouched when it already matches.
func pinGoDirective(m module) error {
	fmt.Printf("==> mingo (%s)\n", m.dir)

	scanner := mingo.Scanner{Deps: true, Indirect: true}
	result, err := scanner.ScanDir(m.dir)
	if err != nil {
		return errors.WithStack(fmt.Errorf("mingo: %w", err))
	}

	declared, err := readGoVersion(m.goMod)
	if err != nil {
		return err
	}

	minimum := "1." + strconv.Itoa(result.Version())
	if declared == minimum {
		fmt.Printf("    go %s is already minimal\n", declared)

		return nil
	}

	fmt.Printf("    go %s -> go %s\n", declared, minimum)

	return writeGoVersion(m, minimum)
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

func writeGoVersion(m module, version string) error {
	file, err := parseModFile(m.goMod)
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
	info, err := os.Stat(m.goMod)
	if err != nil {
		return err
	}

	return os.WriteFile(m.goMod, data, info.Mode().Perm())
}

func parseModFile(path string) (*modfile.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse go.mod: %w", err))
	}

	return file, nil
}
