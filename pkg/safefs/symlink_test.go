package safefs_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/safefs"
)

func TestCheckSymlinks(t *testing.T) {
	t.Parallel()

	type args struct {
		setup    func(t *testing.T, root, skill string)
		srcDir   string
		skillRel string
	}

	tests := []struct {
		name       string
		args       args
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (no symlinks)",
			args: args{
				setup: func(t *testing.T, root, skill string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(skill, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(filepath.Join(skill, "a.txt"), []byte("a"), 0o644)).OrFatal(t)
				},
				srcDir:   "skills",
				skillRel: "skills/skill",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (inner link)",
			args: args{
				setup: func(t *testing.T, root, skill string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(skill, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(filepath.Join(skill, "a.txt"), []byte("a"), 0o644)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("a.txt", filepath.Join(skill, "link"))).OrFatal(t)
				},
				srcDir:   "skills",
				skillRel: "skills/skill",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (loop is skipped)",
			args: args{
				setup: func(t *testing.T, root, skill string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(skill, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(filepath.Join(skill, "b"), filepath.Join(skill, "a"))).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(filepath.Join(skill, "a"), filepath.Join(skill, "b"))).OrFatal(t)
				},
				srcDir:   "skills",
				skillRel: "skills/skill",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (escaping link)",
			args: args{
				setup: func(t *testing.T, root, skill string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(skill, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(filepath.Join(root, "outside.txt"), []byte("x"), 0o644)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(filepath.Join(root, "outside.txt"), filepath.Join(skill, "link"))).OrFatal(t)
				},
				srcDir:   "skills",
				skillRel: "skills/skill",
			},
			errMatcher: internal.ErrorContaining("symlink escapes"),
		},
		{
			name: "fail (missing skill)",
			args: args{
				srcDir:   "skills",
				skillRel: "skills/missing",
			},
			errMatcher: internal.ErrorContaining("missing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			skill := filepath.Join(root, filepath.FromSlash(tt.args.skillRel))
			if tt.args.setup != nil {
				tt.args.setup(t, root, skill)
			}

			tt.errMatcher.Match(safefs.CheckSymlinks(root, tt.args.srcDir, skill)).OrError(t)
		})
	}
}

func TestResolveAllowed(t *testing.T) {
	t.Parallel()

	type args struct {
		setup  func(t *testing.T, root string)
		srcDir string
	}

	type want struct {
		// A nil resolved expects the allowed path back.
		resolved func(t *testing.T, root string) string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "plain dir",
			args: args{
				srcDir: "skills",
			},
			want: want{},
		},
		{
			name: "symlinked dir resolves",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					real := filepath.Join(root, "real")
					its.Nil[error]().Match(os.MkdirAll(real, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(real, filepath.Join(root, "link"))).OrFatal(t)
				},
				srcDir: "link",
			},
			want: want{
				resolved: func(_ *testing.T, root string) string {
					return filepath.Join(root, "real")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root, err := filepath.EvalSymlinks(t.TempDir())
			its.Nil[error]().Match(err).OrFatal(t)

			if tt.args.setup != nil {
				tt.args.setup(t, root)
			}

			allowed, resolved := safefs.ResolveAllowed(root, tt.args.srcDir)
			its.EqEq(filepath.Join(root, tt.args.srcDir)).Match(allowed).OrError(t)

			want := allowed
			if tt.want.resolved != nil {
				want = tt.want.resolved(t, root)
			}
			its.EqEq(want).Match(resolved).OrError(t)
		})
	}
}

func TestResolveTarget(t *testing.T) {
	t.Parallel()

	type args struct {
		setup func(t *testing.T, dir string) string
	}

	tests := []struct {
		name       string
		args       args
		wantOK     bool
		wantTarget func(t *testing.T, dir, resolved string) string
		skipIfRoot bool
	}{
		{
			name: "regular file",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					p := filepath.Join(dir, "a.txt")
					its.Nil[error]().Match(os.WriteFile(p, []byte("a"), 0o644)).OrFatal(t)

					return p
				},
			},
			wantOK: true,
			wantTarget: func(_ *testing.T, dir, resolved string) string {
				return resolved
			},
		},
		{
			name: "inner symlink",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					p := filepath.Join(dir, "a.txt")
					its.Nil[error]().Match(os.WriteFile(p, []byte("a"), 0o644)).OrFatal(t)
					link := filepath.Join(dir, "link")
					its.Nil[error]().Match(os.Symlink(p, link)).OrFatal(t)

					return link
				},
			},
			wantOK: true,
			wantTarget: func(_ *testing.T, dir, _ string) string {
				return filepath.Join(dir, "a.txt")
			},
		},
		{
			name: "dangling link keeps cleaned path",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					link := filepath.Join(dir, "dangling")
					its.Nil[error]().Match(os.Symlink(filepath.Join(dir, "missing"), link)).OrFatal(t)

					return link
				},
			},
			wantOK: true,
			wantTarget: func(_ *testing.T, dir, _ string) string {
				return filepath.Join(dir, "missing")
			},
		},
		{
			name: "loop returns false",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					a := filepath.Join(dir, "a")
					b := filepath.Join(dir, "b")
					its.Nil[error]().Match(os.Symlink(b, a)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(a, b)).OrFatal(t)

					return a
				},
			},
			wantOK: false,
			wantTarget: func(_ *testing.T, dir, _ string) string {
				return ""
			},
		},
		{
			name: "long chain returns false",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					real := filepath.Join(dir, "real.txt")
					its.Nil[error]().Match(os.WriteFile(real, []byte("x"), 0o644)).OrFatal(t)

					prev := real
					for i := range 300 {
						p := filepath.Join(dir, "chain"+strconv.Itoa(i))
						its.Nil[error]().Match(os.Symlink(prev, p)).OrFatal(t)
						prev = p
					}

					return prev
				},
			},
			wantOK: false,
			wantTarget: func(_ *testing.T, dir, _ string) string {
				return ""
			},
		},
		{
			name: "relative link in mid chain",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					a := filepath.Join(dir, "a")
					b := filepath.Join(dir, "b")
					its.Nil[error]().Match(os.Symlink("b", a)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink(filepath.Join(dir, "missing"), b)).OrFatal(t)

					return a
				},
			},
			wantOK: true,
			wantTarget: func(_ *testing.T, dir, _ string) string {
				return filepath.Join(dir, "missing")
			},
		},
		{
			name: "non-symlink hop keeps cleaned path",
			args: args{
				setup: func(t *testing.T, dir string) string {
					t.Helper()
					sub := filepath.Join(dir, "sub")
					its.Nil[error]().Match(os.Mkdir(sub, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Chmod(sub, 0o000)).OrFatal(t)
					t.Cleanup(func() {
						_ = os.Chmod(sub, 0o755)
					})

					link := filepath.Join(dir, "link")
					its.Nil[error]().Match(os.Symlink("sub", link)).OrFatal(t)

					return link
				},
			},
			wantOK: true,
			wantTarget: func(_ *testing.T, dir, _ string) string {
				return filepath.Join(dir, "sub")
			},
			skipIfRoot: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.skipIfRoot && os.Geteuid() == 0 {
				t.Skip("root bypasses permissions")
			}

			dir, err := filepath.EvalSymlinks(t.TempDir())
			its.Nil[error]().Match(err).OrFatal(t)

			p := tt.args.setup(t, dir)

			target, ok := safefs.ResolveTarget(p)
			its.EqEq(tt.wantOK).Match(ok).OrError(t)
			its.EqEq(tt.wantTarget(t, dir, p)).Match(target).OrError(t)
		})
	}

	// Lines 71-72, 75, 80, and 89 need the filesystem to change between the
	// entry EvalSymlinks and each loop step, so they stay uncovered by design.
}

func TestIsWithin(t *testing.T) {
	t.Parallel()

	type args struct {
		target          string
		allowed         string
		allowedResolved string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal allowed",
			args: args{target: "/r/skills", allowed: "/r/skills", allowedResolved: "/real/skills"},
			want: true,
		},
		{
			name: "equal resolved",
			args: args{target: "/real/skills", allowed: "/r/skills", allowedResolved: "/real/skills"},
			want: true,
		},
		{
			name: "under allowed",
			args: args{target: "/r/skills/a", allowed: "/r/skills", allowedResolved: "/real/skills"},
			want: true,
		},
		{
			name: "under resolved",
			args: args{target: "/real/skills/a", allowed: "/r/skills", allowedResolved: "/real/skills"},
			want: true,
		},
		{
			name: "escapes",
			args: args{target: "/r/other", allowed: "/r/skills", allowedResolved: "/real/skills"},
			want: false,
		},
		{
			name: "sibling with shared prefix",
			args: args{target: "/r/skills-extra", allowed: "/r/skills", allowedResolved: "/real/skills"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(safefs.IsWithin(
				tt.args.target,
				tt.args.allowed,
				tt.args.allowedResolved,
			)).OrError(t)
		})
	}
}
