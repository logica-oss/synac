// Package e2e_test exercises the synac binary end to end.
package e2e_test

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cp "github.com/otiai10/copy"
	"github.com/youta-t/its"

	"github.com/logica-oss/synac/pkg/config"
)

const (
	e2eEnv    = "SYNAC_E2E"
	binEnv    = "SYNAC_E2E_BIN"
	envPrefix = "SYNAC_"
)

// TestMain builds the synac binary once for the e2e run.
func TestMain(m *testing.M) {
	if os.Getenv(e2eEnv) != "1" {
		os.Exit(m.Run())
	}

	dir, err := os.MkdirTemp("", "synac-e2e-")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}

	bin := filepath.Join(dir, "synac")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	//nolint:gosec // test-only build of the module binary into a fresh temp dir
	build := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "github.com/logica-oss/synac/cmd/synac")
	if out, err := build.CombinedOutput(); err != nil {
		log.Fatalf("build synac: %v\n%s", err, out)
	}

	if err := os.Setenv(binEnv, bin); err != nil {
		log.Fatalf("set %s: %v", binEnv, err)
	}

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func copyFixture(t *testing.T, fixture, root string) {
	t.Helper()

	its.Nil[error]().Match(cp.Copy(filepath.Join("testdata", fixture), root)).OrFatal(t)
}

func synacBin(t *testing.T) string {
	t.Helper()

	bin := os.Getenv(binEnv)
	if bin == "" {
		t.Fatalf("%s is not set", binEnv)
	}

	return bin
}

func commandEnv(extra map[string]string) []string {
	out := make([]string, 0, len(os.Environ())+len(extra))
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, envPrefix) {
			continue
		}
		out = append(out, entry)
	}

	for key, value := range extra {
		out = append(out, key+"="+value)
	}

	return out
}

type result struct {
	stdout string
	stderr string
	code   int
}

func runSynac(t *testing.T, root string, env map[string]string, args ...string) result {
	t.Helper()

	var stdout, stderr strings.Builder
	//nolint:gosec // test-only execution of the binary built in TestMain
	cmd := exec.CommandContext(t.Context(), synacBin(t), args...)
	cmd.Dir = root
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = commandEnv(env)

	code := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			code = exitErr.ExitCode()
		} else {
			its.Nil[error]().Match(err).OrFatal(t)
		}
	}

	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func TestE2E(t *testing.T) {
	t.Parallel()

	if os.Getenv(e2eEnv) != "1" {
		t.Skip("set SYNAC_E2E=1 to run e2e")
	}

	type args struct {
		fixture string
		argv    func(root string) []string
		env     map[string]string
	}

	type want struct {
		code    int
		files   map[string][]string
		missing []string
		output  string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success (all categories)",
			args: args{
				fixture: "all-categories",
				argv:    func(root string) []string { return []string{"--root", root} },
			},
			want: want{
				code: 0,
				files: map[string][]string{
					"AGENTS.md":                    {"Generated mirror", "project-wide content"},
					".claude/rules/foo.md":         {"Generated from", "path-specific content"},
					".claude/skills/demo/SKILL.md": {"Generated from", "skills content"},
				},
			},
		},
		{
			name: "success (dry-run writes nothing)",
			args: args{
				fixture: "dry-run",
				argv: func(root string) []string {
					return []string{"--root", root, "--dry-run", "--path-specific-source", config.SourceOff, "--skills-source", config.SourceOff}
				},
			},
			want: want{
				code:    0,
				missing: []string{"AGENTS.md"},
			},
		},
		{
			name: "success (all off)",
			args: args{
				fixture: "all-categories",
				argv: func(root string) []string {
					return []string{"--root", root, "--project-wide-source", config.SourceOff, "--path-specific-source", config.SourceOff, "--skills-source", config.SourceOff}
				},
			},
			want: want{
				code: 0,
				missing: []string{
					"AGENTS.md",
					".claude/rules/foo.md",
					".claude/skills/demo/SKILL.md",
				},
			},
		},
		{
			name: "fail (bad instruction)",
			args: args{
				fixture: "bad-instruction",
				argv: func(root string) []string {
					return []string{"--root", root, "--project-wide-source", config.SourceOff, "--path-specific-source", config.SourceGithub, "--skills-source", config.SourceOff}
				},
			},
			want: want{
				code:   1,
				output: "parse bad.instructions.md",
			},
		},
		{
			name: "fail (broken config)",
			args: args{
				fixture: "broken-config",
				argv:    func(root string) []string { return []string{"--root", root} },
			},
			want: want{
				code:   1,
				output: "read config",
			},
		},
		{
			name: "success (flag wins over env and config)",
			args: args{
				fixture: "precedence",
				argv: func(root string) []string {
					return []string{"--root", root, "--project-wide-source", config.SourceGithub}
				},
				env: map[string]string{"SYNAC_PROJECT_WIDE_SOURCE": config.SourceOff},
			},
			want: want{
				code:  0,
				files: map[string][]string{"AGENTS.md": {"Generated mirror", "project-wide content"}},
			},
		},
		{
			name: "success (env wins over config)",
			args: args{
				fixture: "precedence",
				argv:    func(root string) []string { return []string{"--root", root} },
				env:     map[string]string{"SYNAC_PROJECT_WIDE_SOURCE": config.SourceGithub},
			},
			want: want{
				code:  0,
				files: map[string][]string{"AGENTS.md": {"Generated mirror", "project-wide content"}},
			},
		},
		{
			name: "success (config disables all)",
			args: args{
				fixture: "precedence",
				argv:    func(root string) []string { return []string{"--root", root} },
			},
			want: want{
				code:    0,
				missing: []string{"AGENTS.md"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.args.fixture != "" {
				copyFixture(t, tt.args.fixture, root)
			}

			got := runSynac(t, root, tt.args.env, tt.args.argv(root)...)
			its.EqEq(tt.want.code).Match(got.code).OrError(t)
			if tt.want.output != "" {
				its.StringContaining(tt.want.output).Match(got.stdout + got.stderr).OrError(t)
			}

			for rel, wantSubs := range tt.want.files {
				content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
				its.Nil[error]().Match(err).OrError(t)
				for _, wantSub := range wantSubs {
					its.StringContaining(wantSub).Match(string(content)).OrError(t)
				}
			}

			for _, rel := range tt.want.missing {
				_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
				its.EqEq(true).Match(os.IsNotExist(err)).OrError(t)
			}
		})
	}
}
