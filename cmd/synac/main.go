// Command synac synchronizes agent configuration files.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/sync"
)

func newLogger(format string, errWriter bool) *slog.Logger {
	writer := os.Stdout
	if errWriter {
		writer = os.Stderr
	}
	if strings.ToLower(format) == "json" {
		return slog.New(slog.NewJSONHandler(writer, nil))
	}

	return slog.New(slog.NewTextHandler(writer, nil))
}

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "synac: %v\n", err)
		os.Exit(1)
	}

	log := newLogger(cfg.LogFormat, false)
	errLog := newLogger(cfg.LogFormat, true)

	log.Info("starting sync",
		"root", cfg.Root,
		"dry_run", cfg.DryRun,
		"log_format", cfg.LogFormat,
		"project_wide_source", cfg.ProjectWideSource,
		"path_specific_source", cfg.PathSpecificSource,
		"skills_source", cfg.SkillsSource,
	)

	if err := sync.Run(log, cfg); err != nil {
		errLog.Error("sync failed", "error", err)
		os.Exit(1)
	}
}
