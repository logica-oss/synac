package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/config"
)

// Each case must start from defaults because viper reads the environment.
func clearSynacEnv(t *testing.T) {
	t.Helper()

	for _, k := range []string{
		"SYNAC_ROOT",
		"SYNAC_DRY_RUN",
		"SYNAC_LOG_FORMAT",
		"SYNAC_PROJECT_WIDE_SOURCE",
		"SYNAC_PATH_SPECIFIC_SOURCE",
		"SYNAC_SKILLS_SOURCE",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad(t *testing.T) {
	type args struct {
		files map[string]string
		args  func(root string) []string
		// A value of "root" stands for the fresh temp dir.
		env map[string]string
	}

	type want struct {
		dryRun             bool
		logFormat          string
		projectWideSource  string
		pathSpecificSource string
		skillsSource       string
	}

	tests := []struct {
		name       string
		args       args
		want       want
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (defaults with explicit root)",
			args: args{
				args: func(root string) []string { return []string{"--root", root} },
			},
			want: want{
				logFormat:          "console",
				projectWideSource:  "github",
				pathSpecificSource: "github",
				skillsSource:       "agents",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (flags win)",
			args: args{
				args: func(root string) []string {
					return []string{
						"--root", root,
						"--dry-run",
						"--log-format", "json",
						"--project-wide-source", "off",
						"--path-specific-source", "off",
						"--skills-source", "off",
					}
				},
			},
			want: want{
				dryRun:             true,
				logFormat:          "json",
				projectWideSource:  "off",
				pathSpecificSource: "off",
				skillsSource:       "off",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (config file)",
			args: args{
				files: map[string]string{
					".synac.yaml": "dry-run: true\nlog-format: \"json\"\nproject-wide-source: \"off\"\npath-specific-source: \"off\"\nskills-source: \"off\"\n",
				},
				args: func(root string) []string { return []string{"--root", root} },
			},
			want: want{
				dryRun:             true,
				logFormat:          "json",
				projectWideSource:  "off",
				pathSpecificSource: "off",
				skillsSource:       "off",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (explicit config file)",
			args: args{
				files: map[string]string{"custom.yaml": "dry-run: true\n"},
				args: func(root string) []string {
					return []string{"--root", root, "--config", filepath.Join(root, "custom.yaml")}
				},
			},
			want: want{
				dryRun:             true,
				logFormat:          "console",
				projectWideSource:  "github",
				pathSpecificSource: "github",
				skillsSource:       "agents",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (env root)",
			args: args{
				args: func(string) []string { return []string{} },
				env:  map[string]string{"SYNAC_ROOT": "root"},
			},
			want: want{
				logFormat:          "console",
				projectWideSource:  "github",
				pathSpecificSource: "github",
				skillsSource:       "agents",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (root from GITHUB_WORKSPACE)",
			args: args{
				args: func(string) []string { return []string{} },
				env:  map[string]string{"GITHUB_WORKSPACE": "root"},
			},
			want: want{
				logFormat:          "console",
				projectWideSource:  "github",
				pathSpecificSource: "github",
				skillsSource:       "agents",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (help)",
			args: args{
				args: func(string) []string { return []string{"--help"} },
				env:  map[string]string{"SYNAC_ROOT": "root"},
			},
			errMatcher: its.Not(its.Nil[error]()),
		},
		{
			name: "fail (bad flag)",
			args: args{
				args: func(string) []string { return []string{"--no-such-flag"} },
			},
			errMatcher: its.Not(its.Nil[error]()),
		},
		{
			name: "fail (root key in file)",
			args: args{
				files: map[string]string{".synac.yaml": "root: \"/tmp\"\n"},
				args:  func(root string) []string { return []string{"--root", root} },
			},
			errMatcher: internal.ErrorContaining("unsupported config key"),
		},
		{
			name: "fail (invalid value)",
			args: args{
				args: func(root string) []string { return []string{"--root", root, "--log-format", "yaml"} },
			},
			errMatcher: internal.ErrorContaining("invalid log-format"),
		},
		{
			name: "fail (missing root)",
			args: args{
				args: func(string) []string {
					return []string{"--root", filepath.Join(t.TempDir(), "missing")}
				},
			},
			errMatcher: internal.ErrorContaining("resolve root"),
		},
		{
			name: "fail (broken config)",
			args: args{
				files: map[string]string{".synac.yaml": "foo: [bar\n"},
				args:  func(root string) []string { return []string{"--root", root} },
			},
			errMatcher: internal.ErrorContaining("read config"),
		},
		{
			name: "fail (undecodable config)",
			args: args{
				files: map[string]string{".synac.yaml": "dry-run: notabool\n"},
				args:  func(root string) []string { return []string{"--root", root} },
			},
			errMatcher: internal.ErrorContaining("decode config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			clearSynacEnv(t)
			t.Setenv("GITHUB_WORKSPACE", "")

			for rel, content := range tt.args.files {
				p := filepath.Join(root, filepath.FromSlash(rel))
				its.Nil[error]().Match(os.MkdirAll(filepath.Dir(p), 0o755)).OrFatal(t)
				its.Nil[error]().Match(os.WriteFile(p, []byte(content), 0o644)).OrFatal(t)
			}

			for k, v := range tt.args.env {
				if v == "root" {
					v = root
				}
				t.Setenv(k, v)
			}

			cfg, err := config.Load(tt.args.args(root))
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.EqEq(root).Match(cfg.Root).OrError(t)
			its.EqEq(tt.want.dryRun).Match(cfg.DryRun).OrError(t)
			its.EqEq(tt.want.logFormat).Match(cfg.LogFormat).OrError(t)
			its.EqEq(tt.want.projectWideSource).Match(cfg.ProjectWideSource).OrError(t)
			its.EqEq(tt.want.pathSpecificSource).Match(cfg.PathSpecificSource).OrError(t)
			its.EqEq(tt.want.skillsSource).Match(cfg.SkillsSource).OrError(t)
		})
	}
}
