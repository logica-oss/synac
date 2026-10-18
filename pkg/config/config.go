// Package config loads synac settings.
package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Source values for sync origins.
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

// Config holds synac settings.
type Config struct {
	Root               string `mapstructure:"-"` // Stays out of the config file to keep relative resolution unambiguous.
	Version            bool   `mapstructure:"-"` // Stays out of the config file because it is a CLI-only flag.
	DryRun             bool   `mapstructure:"dry-run"`
	LogFormat          string `mapstructure:"log-format"`
	ProjectWideSource  string `mapstructure:"project-wide-source"`
	PathSpecificSource string `mapstructure:"path-specific-source"`
	SkillsSource       string `mapstructure:"skills-source"`
}

const (
	keyRoot               = "root"
	keyVersion            = "version"
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
	flagSet := pflag.NewFlagSet("synac", pflag.ContinueOnError)

	flagSet.StringP(keyConfig, "c", "", "config file path (default: <root>/.synac.yaml or .synac.json)")
	flagSet.String(keyRoot, "", "repository root (default: $GITHUB_WORKSPACE or git top level)")
	flagSet.BoolP(keyVersion, "v", false, "print version and exit")
	flagSet.BoolP(keyDryRun, "n", false, "print planned changes without writing")
	flagSet.String(keyLogFormat, logFormatConsole, "log format: console or json")
	flagSet.String(keyProjectWideSource, SourceGithub, "project-wide source: github, agents, or off")
	flagSet.String(keyPathSpecificSource, SourceGithub, "path-specific source: github, claude, or off")
	flagSet.String(keySkillsSource, SourceAgents, "skills source: agents, claude, or off")

	flagSet.Usage = func() {
		_, _ = fmt.Fprintf(flagSet.Output(),
			"Usage: synac [options]\n\nSync agent configs from canonical sources.\n\nOptions:\n")
		flagSet.PrintDefaults()
		_, _ = fmt.Fprintf(flagSet.Output(), "\nConfig file: .synac.yaml (preferred, supports comments)"+
			" or .synac.json in the repository root.\nEnvironment: SYNAC_* variables"+
			" (e.g. SYNAC_DRY_RUN, SYNAC_LOG_FORMAT).\n"+
			"Precedence: flag > env > config file > default.\n")
	}

	return flagSet
}
