package config_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/config"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	valid := config.Config{
		LogFormat:          "console",
		ProjectWideSource:  "github",
		PathSpecificSource: "github",
		SkillsSource:       "agents",
	}

	type args struct {
		cfg config.Config
	}

	tests := []struct {
		name       string
		args       args
		errMatcher its.Matcher[error]
	}{
		{
			name:       "success (console)",
			args:       args{cfg: valid},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (json and off)",
			args: args{cfg: config.Config{
				LogFormat:          "json",
				ProjectWideSource:  "off",
				PathSpecificSource: "off",
				SkillsSource:       "off",
			}},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (agents and claude)",
			args: args{cfg: config.Config{
				LogFormat:          "console",
				ProjectWideSource:  "agents",
				PathSpecificSource: "claude",
				SkillsSource:       "claude",
			}},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (log format)",
			args: args{cfg: config.Config{
				LogFormat:          "yaml",
				ProjectWideSource:  valid.ProjectWideSource,
				PathSpecificSource: valid.PathSpecificSource,
				SkillsSource:       valid.SkillsSource,
			}},
			errMatcher: internal.ErrorContaining("invalid log-format"),
		},
		{
			name: "fail (project-wide)",
			args: args{cfg: config.Config{
				LogFormat:          valid.LogFormat,
				ProjectWideSource:  "claude",
				PathSpecificSource: valid.PathSpecificSource,
				SkillsSource:       valid.SkillsSource,
			}},
			errMatcher: internal.ErrorContaining("invalid project-wide-source"),
		},
		{
			name: "fail (path-specific)",
			args: args{cfg: config.Config{
				LogFormat:          valid.LogFormat,
				ProjectWideSource:  valid.ProjectWideSource,
				PathSpecificSource: "agents",
				SkillsSource:       valid.SkillsSource,
			}},
			errMatcher: internal.ErrorContaining("invalid path-specific-source"),
		},
		{
			name: "fail (skills)",
			args: args{cfg: config.Config{
				LogFormat:          valid.LogFormat,
				ProjectWideSource:  valid.ProjectWideSource,
				PathSpecificSource: valid.PathSpecificSource,
				SkillsSource:       "github",
			}},
			errMatcher: internal.ErrorContaining("invalid skills-source"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.errMatcher.Match(config.Validate(tt.args.cfg)).OrError(t)
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()

	type args struct {
		values []string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{values: nil},
			want: "",
		},
		{
			name: "all empty",
			args: args{values: []string{"", ""}},
			want: "",
		},
		{
			name: "first wins",
			args: args{values: []string{"a", "b"}},
			want: "a",
		},
		{
			name: "skips empty",
			args: args{values: []string{"", "b", "c"}},
			want: "b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(config.FirstNonEmpty(tt.args.values...)).OrError(t)
		})
	}
}

func TestPreferredConfigFile(t *testing.T) {
	t.Parallel()

	type args struct {
		files    []string
		rootOnly bool
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty when missing",
			args: args{rootOnly: true},
			want: "",
		},
		{
			name: "prefers yaml in root",
			args: args{files: []string{".synac.yaml", ".synac.json"}, rootOnly: true},
			want: ".synac.yaml",
		},
		{
			name: "falls back to json",
			args: args{files: []string{".synac.json"}},
			want: ".synac.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for _, name := range tt.args.files {
				body := []byte("{}")
				if name == ".synac.yaml" {
					body = []byte("dry-run: true\n")
				}
				its.Nil[error]().Match(os.WriteFile(filepath.Join(dir, name), body, 0o644)).OrFatal(t)
			}

			got := config.PreferredConfigFile(dir, tt.args.rootOnly)
			if tt.want == "" {
				its.EqEq("").Match(got).OrError(t)

				return
			}
			its.EqEq(filepath.Join(dir, tt.want)).Match(got).OrError(t)
		})
	}
}

func TestResolveRoot(t *testing.T) {
	type args struct {
		root func(t *testing.T, dir string) string
		// A value of "dir" or "missing" stands for a path under the fresh temp dir.
		env map[string]string
		// wantCwd expects the fallback to the working directory.
		wantCwd bool
	}

	tests := []struct {
		name       string
		args       args
		errMatcher its.Matcher[error]
	}{
		{
			name:       "success (explicit dir)",
			args:       args{root: func(_ *testing.T, dir string) string { return dir }},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (detected git top level)",
			args: args{
				root: func(_ *testing.T, _ string) string { return "" },
				env:  map[string]string{"GITHUB_WORKSPACE": "dir"},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (falls back to cwd)",
			args: args{
				root:    func(_ *testing.T, _ string) string { return "" },
				env:     map[string]string{"GIT_DIR": "missing"},
				wantCwd: true,
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (missing)",
			args: args{
				root: func(t *testing.T, dir string) string {
					t.Helper()

					return filepath.Join(dir, "missing")
				},
			},
			errMatcher: internal.ErrorContaining("resolve root"),
		},
		{
			name: "fail (not a dir)",
			args: args{
				root: func(t *testing.T, dir string) string {
					t.Helper()
					f := filepath.Join(dir, "f")
					its.Nil[error]().Match(os.WriteFile(f, []byte("x"), 0o644)).OrFatal(t)

					return f
				},
			},
			errMatcher: internal.ErrorContaining("not a directory"),
		},
		{
			name: "fail (cwd removed)",
			args: args{
				root: func(t *testing.T, dir string) string {
					t.Helper()
					sub := filepath.Join(dir, "gone")
					its.Nil[error]().Match(os.Mkdir(sub, 0o755)).OrFatal(t)
					t.Chdir(sub)
					its.Nil[error]().Match(os.Remove(sub)).OrFatal(t)

					return ""
				},
				env: map[string]string{"GIT_DIR": "missing"},
			},
			errMatcher: internal.ErrorContaining("resolve root"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_WORKSPACE", "")
			t.Setenv("GIT_DIR", "")

			dir := t.TempDir()
			for key, value := range tt.args.env {
				switch value {
				case "dir":
					value = dir

				case "missing":
					value = filepath.Join(dir, "missing")
				}
				t.Setenv(key, value)
			}

			want := dir
			if tt.args.wantCwd {
				// Chdir back to a live directory so Getwd works for both sides.
				t.Chdir(dir)
				cwd, err := os.Getwd()
				its.Nil[error]().Match(err).OrFatal(t)
				want = cwd
			}

			got, err := config.ResolveRoot(tt.args.root(t, dir))
			tt.errMatcher.Match(err).OrError(t)
			if err != nil {
				return
			}

			its.EqEq(want).Match(got).OrError(t)
		})
	}

	// filepath.Abs fails only when Getwd fails on a relative path, but Getwd
	// already returned above, so that branch stays uncovered by design.
}

func TestDetectRoot(t *testing.T) {
	type args struct {
		// A value of "dir" or "missing" stands for a path under the fresh temp dir.
		env map[string]string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "workspace env wins",
			args: args{env: map[string]string{"GITHUB_WORKSPACE": "dir"}},
			want: "dir",
		},
		{
			name: "git toplevel when no env",
			args: args{env: map[string]string{}},
			want: "git",
		},
		{
			name: "empty when git fails",
			args: args{env: map[string]string{"GIT_DIR": "missing"}},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_WORKSPACE", "")

			dir := t.TempDir()
			for key, value := range tt.args.env {
				switch value {
				case "dir":
					value = dir

				case "missing":
					value = filepath.Join(dir, "missing")
				}
				t.Setenv(key, value)
			}
			if _, ok := tt.args.env["GIT_DIR"]; !ok {
				_ = os.Unsetenv("GIT_DIR")
			}

			want := tt.want
			switch want {
			case "dir":
				want = dir

			case "git":
				out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel").Output()
				its.Nil[error]().Match(err).OrFatal(t)
				want = strings.TrimSpace(string(out))
			}

			its.EqEq(want).Match(config.DetectRoot()).OrError(t)
		})
	}
}
