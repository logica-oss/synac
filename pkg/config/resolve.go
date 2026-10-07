package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/k1LoW/errors"
)

// Validate checks source selections and log format.
func Validate(cfg Config) error {
	if cfg.LogFormat != logFormatConsole && cfg.LogFormat != logFormatJSON {
		return errors.WithStack(fmt.Errorf("invalid log-format %q: must be %q or %q",
			cfg.LogFormat, logFormatConsole, logFormatJSON))
	}

	switch cfg.ProjectWideSource {
	case SourceGithub, SourceAgents, SourceOff:

	default:
		return errors.WithStack(fmt.Errorf("invalid project-wide-source %q: must be github, agents, or off",
			cfg.ProjectWideSource))
	}

	switch cfg.PathSpecificSource {
	case SourceGithub, SourceClaude, SourceOff:

	default:
		return errors.WithStack(fmt.Errorf("invalid path-specific-source %q: must be github, claude, or off",
			cfg.PathSpecificSource))
	}

	switch cfg.SkillsSource {
	case SourceAgents, SourceClaude, SourceOff:

	default:
		return errors.WithStack(fmt.Errorf("invalid skills-source %q: must be agents, claude, or off", cfg.SkillsSource))
	}

	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

func preferredConfigFile(root string, rootOnly bool) string {
	dirs := []string{"."}
	if root != "" {
		dirs = []string{root, "."}
		if rootOnly {
			dirs = []string{root}
		}
	}

	for _, dir := range dirs {
		for _, name := range []string{".synac.yaml", ".synac.json"} {
			p := filepath.Join(dir, name)

			//nolint:gosec // dir is "." or the resolved root, name is a fixed config filename
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}

	return ""
}

func resolveRoot(root string) (string, error) {
	if root == "" {
		root = detectRoot()
	}

	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.WithStack(fmt.Errorf("resolve root: %w", err))
		}

		root = cwd
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("resolve root: %w", err))
	}

	st, err := os.Stat(abs)
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("resolve root: %w", err))
	}
	if !st.IsDir() {
		return "", errors.WithStack(fmt.Errorf("resolve root %q: not a directory", root))
	}

	return abs, nil
}

func detectRoot() string {
	if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
		return ws
	}

	out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
