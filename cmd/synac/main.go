// Command synac synchronizes agent configuration files.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/pflag"

	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/sync"
)

// version is injected at build time via ldflags.
//
//nolint:gochecknoglobals // build-time injection requires a package variable.
var version = "dev"

func newLogger(format string, errWriter bool) *slog.Logger {
	writer := os.Stdout
	if errWriter {
		writer = os.Stderr
	}

	switch format {
	case config.LogFormatJSON:
		return slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	case config.LogFormatConsole:
		return slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))

	default:
		return nil
	}
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

	if cfg.Version {
		if _, err := fmt.Fprintln(os.Stdout, version); err != nil {
			fmt.Fprintf(os.Stderr, "synac: print version: %v\n", err)
			os.Exit(1)
		}

		return
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
