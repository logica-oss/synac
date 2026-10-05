// Package config manages synac configuration via viper.
//
// Configuration precedence is flag > environment > config file > default.
// Config files are searched as .synac.yaml (preferred) then .synac.json
// in the repository root. YAML is recommended for human editing because
// it supports comments, JSON is supported for machine generation.
// Environment variables use the SYNAC_ prefix, for example
// SYNAC_DRY_RUN or SYNAC_PROJECT_WIDE_SOURCE.
package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	SourceGithub = "github"
	SourceAgents = "agents"
	SourceClaude = "claude"
	SourceOff    = "off"
)

const (
	logFormatConsole = "console"
	logFormatJSON    = "json"
)

type Config struct {
	// Root is the repository root. It is resolved from --root, SYNAC_ROOT,
	// or auto-detection, never from the config file.
	Root               string `mapstructure:"-"`
	DryRun             bool   `mapstructure:"dry-run"`
	LogFormat          string `mapstructure:"log-format"`
	ProjectWideSource  string `mapstructure:"project-wide-source"`
	PathSpecificSource string `mapstructure:"path-specific-source"`
	SkillsSource       string `mapstructure:"skills-source"`
}

const (
	keyRoot               = "root"
	keyDryRun             = "dry-run"
	keyLogFormat          = "log-format"
	keyProjectWideSource  = "project-wide-source"
	keyPathSpecificSource = "path-specific-source"
	keySkillsSource       = "skills-source"
	keyConfig             = "config"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault(keyDryRun, false)
	v.SetDefault(keyLogFormat, logFormatConsole)
	v.SetDefault(keyProjectWideSource, SourceGithub)
	v.SetDefault(keyPathSpecificSource, SourceGithub)
	v.SetDefault(keySkillsSource, SourceAgents)
}

func flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("synac", pflag.ContinueOnError)

	fs.String(keyConfig, "", "config file path (default: <root>/.synac.yaml or .synac.json)")
	fs.String(keyRoot, "", "repository root (default: $GITHUB_WORKSPACE or git top level)")
	fs.Bool(keyDryRun, false, "print planned changes without writing")
	fs.String(keyLogFormat, logFormatConsole, "log format: console or json")
	fs.String(keyProjectWideSource, SourceGithub, "project-wide source: github, agents, or off")
	fs.String(keyPathSpecificSource, SourceGithub, "path-specific source: github, claude, or off")
	fs.String(keySkillsSource, SourceAgents, "skills source: agents, claude, or off")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: synac [options]\n\nSync agent configs from canonical sources.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nConfig file: .synac.yaml (preferred, supports comments) or .synac.json in the repository root.\nEnvironment: SYNAC_* variables (e.g. SYNAC_DRY_RUN, SYNAC_LOG_FORMAT).\nPrecedence: flag > env > config file > default.\n")
	}

	return fs
}
