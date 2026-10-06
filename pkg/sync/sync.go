// Package sync synchronizes agent configs.
package sync

import (
	"log/slog"

	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/safefs"
)

type runner struct {
	log     *slog.Logger
	root    string
	applier *safefs.Applier
}

// Run synchronizes all agent configs.
func Run(log *slog.Logger, cfg config.Config) error {
	if err := config.Validate(cfg); err != nil {
		return err
	}

	applier, err := safefs.NewApplier(log, cfg.Root, cfg.DryRun)
	if err != nil {
		return err
	}
	defer func() {
		_ = applier.Close()
	}()

	return (&runner{
		log:     log,
		root:    cfg.Root,
		applier: applier,
	}).run(cfg)
}

func (r *runner) run(cfg config.Config) error {
	if err := r.sync(cfg.ProjectWideSource, "project-wide", r.syncProjectWide); err != nil {
		return err
	}

	if err := r.sync(cfg.PathSpecificSource, "path-specific", r.syncPathSpecific); err != nil {
		return err
	}

	if err := r.sync(cfg.SkillsSource, "skills", r.syncSkills); err != nil {
		return err
	}

	r.log.Info("sync summary", "dry_run", cfg.DryRun)

	return nil
}

func (r *runner) sync(source, category string, sync func(string) error) error {
	if source == config.SourceOff {
		r.log.Info("skip sync", "category", category, "source", config.SourceOff)

		return nil
	}

	return sync(source)
}
