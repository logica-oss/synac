package sync_test

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/sync"
	"github.com/youta-t/its"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	its.Nil[error]().Match(os.MkdirAll(filepath.Dir(path), 0o755)).OrFatal(t)
	its.Nil[error]().Match(os.WriteFile(path, []byte(content), 0o644)).OrFatal(t)
}

func TestRun(t *testing.T) {
	type args struct {
		setup func(t *testing.T, root string)
		cfg   func(root string) config.Config
	}

	type want struct {
		files    map[string]string
		missing  []string
		symlinks []string
	}

	tests := []struct {
		name       string
		args       args
		want       want
		skipIfRoot bool
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (all categories)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "copilot-instructions.md"), `# hello
`)
					writeFile(t, filepath.Join(root, ".github", "instructions", "foo.instructions.md"), `---
applyTo: "a"
---
hello
`)
					writeFile(t, filepath.Join(root, ".agents", "skills", "demo", "SKILL.md"), `---
name: demo
---
hello
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "github",
						PathSpecificSource: "github",
						SkillsSource:       "agents",
					}
				},
			},
			want: want{
				files: map[string]string{
					"AGENTS.md":                    "Generated mirror",
					".claude/rules/foo.md":         "Generated from",
					".claude/skills/demo/SKILL.md": "Generated from",
				},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (all off)",
			args: args{
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "off",
					}
				},
			},
			want:       want{},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (missing sources)",
			args: args{
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "agents",
					}
				},
			},
			want:       want{},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (claude path-specific)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".claude", "rules", "foo.md"), `---
paths:
  - "a"
---
hello
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "claude",
						SkillsSource:       "off",
					}
				},
			},
			want: want{
				files: map[string]string{
					".github/instructions/foo.instructions.md": "Generated from",
				},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (skill without SKILL.md)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".agents", "skills", "demo", "other.txt"), `hi
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			want: want{
				files: map[string]string{
					".claude/skills/demo/other.txt": `hi
`,
				},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (dest SKILL.md symlink kept)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					skill := filepath.Join(root, ".agents", "skills", "demo")
					writeFile(t, filepath.Join(skill, "real.md"), `---
name: demo
---
hello
`)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(skill, "real.md"),
						filepath.Join(skill, "SKILL.md"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			want: want{
				symlinks: []string{".claude/skills/demo/SKILL.md"},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (dry-run writes nothing)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "copilot-instructions.md"), `# hello
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						DryRun:             true,
						ProjectWideSource:  "github",
						PathSpecificSource: "off",
						SkillsSource:       "off",
					}
				},
			},
			want: want{
				missing: []string{"AGENTS.md"},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (invalid config)",
			args: args{
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "yaml",
						ProjectWideSource:  "github",
						PathSpecificSource: "github",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("invalid log-format"),
		},
		{
			name: "fail (missing root)",
			args: args{
				cfg: func(string) config.Config {
					return config.Config{
						Root:               filepath.Join(t.TempDir(), "missing"),
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: its.Not(its.Nil[error]()),
		},
		{
			name: "fail (missing project-wide source)",
			args: args{
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "github",
						PathSpecificSource: "off",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("project-wide source not found"),
		},
		{
			name: "fail (bad instruction)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "instructions", "bad.instructions.md"), `---
applyTo: 123
---
hello
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("parse bad.instructions.md"),
		},
		{
			name: "fail (escaping skill symlink)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					skill := filepath.Join(root, ".agents", "skills", "demo")
					writeFile(t, filepath.Join(skill, "SKILL.md"), `hello
`)
					writeFile(t, filepath.Join(root, "outside.txt"), `x
`)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(root, "outside.txt"),
						filepath.Join(skill, "link"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("symlink escapes"),
		},
		{
			name: "fail (path-specific unreadable source)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "instructions"), `not a dir
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("read source dir"),
		},
		{
			name: "fail (path-specific same dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, ".github", "instructions"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, ".claude"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(root, ".github", "instructions"),
						filepath.Join(root, ".claude", "rules"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("path-specific source and destination are the same directory"),
		},
		{
			name: "fail (project-wide same file)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "copilot-instructions.md"), `# hello
`)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(root, ".github", "copilot-instructions.md"),
						filepath.Join(root, "AGENTS.md"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "github",
						PathSpecificSource: "off",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("same file"),
		},
		{
			name: "fail (skills same dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, ".agents", "skills", "demo"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, ".claude"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(root, ".agents", "skills"),
						filepath.Join(root, ".claude", "skills"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("skills source and destination are the same directory"),
		},
		{
			name: "fail (project-wide write to dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "copilot-instructions.md"), `# hello
`)
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "AGENTS.md"), 0o755)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "github",
						PathSpecificSource: "off",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("write AGENTS.md"),
		},
		{
			name: "fail (skills within rejected)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "real"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(root, "real"),
						filepath.Join(root, ".agents"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("symlinked path not allowed"),
		},
		{
			name: "fail (path-specific remove dest)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "instructions", "foo.instructions.md"), `---
applyTo: "a"
---
hello
`)
					writeFile(t, filepath.Join(root, ".claude"), `file
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("validate dest"),
		},
		{
			name: "fail (instruction dangling link)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					dir := filepath.Join(root, ".github", "instructions")
					its.Nil[error]().Match(os.MkdirAll(dir, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(dir, "missing"),
						filepath.Join(dir, "foo.instructions.md"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("read"),
		},
		{
			name: "fail (skills source is file)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".agents", "skills"), `file
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("read source dir"),
		},
		{
			name: "fail (skills remove dest)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".agents", "skills", "demo", "SKILL.md"), `hello
`)
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, ".claude"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(
						filepath.Join(root, "outside"),
						filepath.Join(root, ".claude", "skills"),
					)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("validate dest"),
		},
		{
			name: "fail (skill SKILL.md is dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, ".agents", "skills", "demo", "SKILL.md"), 0o755)).OrFatal(t)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("is a directory"),
		},
		{
			name: "fail (instruction name too long)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".claude", "rules", strings.Repeat("a", 250)+".md"), `---
paths:
  - "a"
---
hello
`)
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "claude",
						SkillsSource:       "off",
					}
				},
			},
			errMatcher: internal.ErrorContaining("file name too long"),
		},
		{
			name: "fail (path-specific dest dir is read-only)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".github", "instructions", "foo.instructions.md"), `---
applyTo: "a"
---
hello
`)
					its.Nil[error]().Match(os.Chmod(root, 0o555)).OrFatal(t)
					t.Cleanup(func() {
						_ = os.Chmod(root, 0o755)
					})
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "github",
						SkillsSource:       "off",
					}
				},
			},
			want:       want{},
			errMatcher: internal.ErrorContaining("create dest dir .claude/rules"),
			skipIfRoot: true,
		},
		{
			name: "fail (skills dest dir is read-only)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					writeFile(t, filepath.Join(root, ".agents", "skills", "demo", "SKILL.md"), `hello
`)
					its.Nil[error]().Match(os.Chmod(root, 0o555)).OrFatal(t)
					t.Cleanup(func() {
						_ = os.Chmod(root, 0o755)
					})
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("create dest dir .claude/skills"),
			skipIfRoot: true,
		},
		{
			name: "fail (skill file is unreadable)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					skill := filepath.Join(root, ".agents", "skills", "demo")
					writeFile(t, filepath.Join(skill, "SKILL.md"), `hello
`)
					writeFile(t, filepath.Join(skill, "secret.txt"), `secret
`)
					its.Nil[error]().Match(os.Chmod(filepath.Join(skill, "secret.txt"), 0o000)).OrFatal(t)
					t.Cleanup(func() {
						_ = os.Chmod(filepath.Join(skill, "secret.txt"), 0o644)
					})
				},
				cfg: func(root string) config.Config {
					return config.Config{
						Root:               root,
						LogFormat:          "console",
						ProjectWideSource:  "off",
						PathSpecificSource: "off",
						SkillsSource:       "agents",
					}
				},
			},
			errMatcher: internal.ErrorContaining("copy skill"),
			skipIfRoot: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfRoot && os.Geteuid() == 0 {
				t.Skip("root bypasses permissions")
			}

			root := t.TempDir()
			if tt.args.setup != nil {
				tt.args.setup(t, root)
			}

			err := sync.Run(newTestLogger(), tt.args.cfg(root))
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			for rel, wantSub := range tt.want.files {
				got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
				its.Nil[error]().Match(err).OrError(t)
				its.StringContaining(wantSub).Match(string(got)).OrError(t)
			}

			for _, rel := range tt.want.missing {
				_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
				its.EqEq(true).Match(os.IsNotExist(err)).OrError(t)
			}

			for _, rel := range tt.want.symlinks {
				fi, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel)))
				its.Nil[error]().Match(err).OrError(t)
				its.EqEq(true).Match(fi.Mode()&os.ModeSymlink != 0).OrError(t)
			}
		})
	}
}
