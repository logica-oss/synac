package safefs_test

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/safefs"
)

func newTestApplier(t *testing.T, dryRun bool) (*safefs.Applier, string) {
	t.Helper()

	root := t.TempDir()
	a, err := safefs.NewApplier(slog.New(slog.NewTextHandler(os.Stderr, nil)), root, dryRun)
	its.Nil[error]().Match(err).OrFatal(t)
	t.Cleanup(func() {
		_ = a.Close()
	})

	return a, root
}

func TestNewApplier(t *testing.T) {
	t.Parallel()

	type args struct {
		dryRun      bool
		missingRoot bool
	}

	tests := []struct {
		name       string
		args       args
		errMatcher its.Matcher[error]
	}{
		{
			name:       "success",
			args:       args{},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "fail (missing root)",
			args:       args{missingRoot: true},
			errMatcher: internal.ErrorContaining("open root"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.args.missingRoot {
				root = filepath.Join(root, "missing")
			}

			a, err := safefs.NewApplier(
				slog.New(slog.NewTextHandler(os.Stderr, nil)),
				root,
				tt.args.dryRun,
			)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.Nil[error]().Match(a.Close()).OrError(t)
		})
	}
}

func TestWithin(t *testing.T) {
	t.Parallel()

	type args struct {
		setup   func(t *testing.T, root string)
		destRel string
	}

	tests := []struct {
		name       string
		args       args
		errMatcher its.Matcher[error]
	}{
		{
			name:       "success (dot)",
			args:       args{destRel: "."},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "success (empty)",
			args:       args{destRel: ""},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "success (missing prefix is allowed)",
			args:       args{destRel: "new/dir"},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (existing dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "ok"), 0o755)).OrFatal(t)
				},
				destRel: "ok/file.txt",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (dot parts skipped)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "ok"), 0o755)).OrFatal(t)
				},
				destRel: "./ok/./file.txt",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (symlinked prefix)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "real"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("real", filepath.Join(root, "link"))).OrFatal(t)
				},
				destRel: "link/file.txt",
			},
			errMatcher: internal.ErrorContaining("symlinked path not allowed"),
		},
		{
			name:       "fail (escaping root)",
			args:       args{destRel: ".."},
			errMatcher: internal.ErrorContaining("validate dest"),
		},
		{
			name: "fail (file as prefix)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.WriteFile(filepath.Join(root, "f"), []byte("x"), 0o644)).OrFatal(t)
				},
				destRel: "f/g.txt",
			},
			errMatcher: internal.ErrorContaining("validate dest"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, root := newTestApplier(t, false)
			if tt.args.setup != nil {
				tt.args.setup(t, root)
			}

			tt.errMatcher.Match(a.Within(tt.args.destRel)).OrError(t)
		})
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()

	type args struct {
		dryRun  bool
		setup   func(t *testing.T, root string)
		destRel string
		data    string
		perm    fs.FileMode
	}

	type want struct {
		content string
		missing bool
	}

	tests := []struct {
		name       string
		args       args
		want       want
		skipIfRoot bool
		errMatcher its.Matcher[error]
	}{
		{
			name:       "success (creates parents)",
			args:       args{destRel: "a/b/c.txt", data: "hello", perm: 0o644},
			want:       want{content: "hello"},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "success (dry-run writes nothing)",
			args:       args{dryRun: true, destRel: "dry/file.txt", data: "hello", perm: 0o644},
			want:       want{missing: true},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (symlinked dest)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "real"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("real", filepath.Join(root, "link"))).OrFatal(t)
				},
				destRel: "link/f.txt",
				data:    "x",
				perm:    0o644,
			},
			errMatcher: internal.ErrorContaining("symlinked path not allowed"),
		},
		{
			name: "fail (parent is file)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.WriteFile(filepath.Join(root, "f"), []byte("x"), 0o644)).OrFatal(t)
				},
				destRel: "f/g.txt",
				data:    "x",
				perm:    0o644,
			},
			errMatcher: internal.ErrorContaining("validate dest"),
		},
		{
			name: "fail (dest is dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "d"), 0o755)).OrFatal(t)
				},
				destRel: "d",
				data:    "x",
				perm:    0o644,
			},
			errMatcher: internal.ErrorContaining("write d"),
		},
		{
			name: "fail (parent not writable)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "ro"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Chmod(filepath.Join(root, "ro"), 0o555)).OrFatal(t)
					t.Cleanup(func() {
						_ = os.Chmod(filepath.Join(root, "ro"), 0o755)
					})
				},
				destRel: "ro/sub/f.txt",
				data:    "x",
				perm:    0o644,
			},
			errMatcher: internal.ErrorContaining("create parent dir"),
			skipIfRoot: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.skipIfRoot && os.Geteuid() == 0 {
				t.Skip("root bypasses permissions")
			}

			a, root := newTestApplier(t, tt.args.dryRun)
			if tt.args.setup != nil {
				tt.args.setup(t, root)
			}

			err := a.WriteFile(
				filepath.Join(root, filepath.FromSlash(tt.args.destRel)),
				tt.args.destRel,
				[]byte(tt.args.data),
				tt.args.perm,
			)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			if tt.want.missing {
				_, err := os.Stat(filepath.Join(root, filepath.FromSlash(tt.args.destRel)))
				its.EqEq(true).Match(os.IsNotExist(err)).OrError(t)

				return
			}

			got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tt.args.destRel)))
			its.Nil[error]().Match(err).OrError(t)
			its.EqEq(tt.want.content).Match(string(got)).OrError(t)
		})
	}
}

func TestMkdirAll(t *testing.T) {
	t.Parallel()

	type args struct {
		dryRun  bool
		setup   func(t *testing.T, root string)
		destRel string
		perm    fs.FileMode
	}

	tests := []struct {
		name       string
		args       args
		missing    bool
		errMatcher its.Matcher[error]
	}{
		{
			name:       "success",
			args:       args{destRel: "a/b", perm: 0o755},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "success (dry-run creates nothing)",
			args:       args{dryRun: true, destRel: "dry/dir", perm: 0o755},
			missing:    true,
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (symlinked dest)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "real"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("real", filepath.Join(root, "link"))).OrFatal(t)
				},
				destRel: "link/sub",
				perm:    0o755,
			},
			errMatcher: internal.ErrorContaining("symlinked path not allowed"),
		},
		{
			name: "fail (existing file)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.WriteFile(filepath.Join(root, "f"), []byte("x"), 0o644)).OrFatal(t)
				},
				destRel: "f",
				perm:    0o755,
			},
			errMatcher: internal.ErrorContaining("create dest dir"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, root := newTestApplier(t, tt.args.dryRun)
			if tt.args.setup != nil {
				tt.args.setup(t, root)
			}

			err := a.MkdirAll(
				filepath.Join(root, filepath.FromSlash(tt.args.destRel)),
				tt.args.destRel,
				tt.args.perm,
			)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			info, err := os.Stat(filepath.Join(root, filepath.FromSlash(tt.args.destRel)))
			if tt.missing {
				its.EqEq(true).Match(os.IsNotExist(err)).OrError(t)

				return
			}

			its.Nil[error]().Match(err).OrError(t)
			its.EqEq(true).Match(info.IsDir()).OrError(t)
		})
	}
}

func TestRemoveAll(t *testing.T) {
	t.Parallel()

	type args struct {
		dryRun  bool
		setup   func(t *testing.T, root string)
		destRel string
	}

	tests := []struct {
		name       string
		args       args
		keepDir    bool
		skipIfRoot bool
		errMatcher its.Matcher[error]
	}{
		{
			name: "success",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "gone"), 0o755)).OrFatal(t)
				},
				destRel: "gone",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (dry-run keeps dir)",
			args: args{
				dryRun: true,
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "keep"), 0o755)).OrFatal(t)
				},
				destRel: "keep",
			},
			keepDir:    true,
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (symlinked dest)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "real"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("real", filepath.Join(root, "link"))).OrFatal(t)
				},
				destRel: "link",
			},
			errMatcher: internal.ErrorContaining("symlinked path not allowed"),
		},
		{
			name: "fail (read-only dir)",
			args: args{
				setup: func(t *testing.T, root string) {
					t.Helper()
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "ro"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(filepath.Join(root, "ro", "keep.txt"), []byte("x"), 0o644)).OrFatal(t)
					its.Nil[error]().Match(os.Chmod(filepath.Join(root, "ro"), 0o555)).OrFatal(t)
					t.Cleanup(func() {
						_ = os.Chmod(filepath.Join(root, "ro"), 0o755)
					})
				},
				destRel: "ro/keep.txt",
			},
			skipIfRoot: true,
			errMatcher: internal.ErrorContaining("clear dest dir"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.skipIfRoot && os.Geteuid() == 0 {
				t.Skip("root bypasses permissions")
			}

			a, root := newTestApplier(t, tt.args.dryRun)
			if tt.args.setup != nil {
				tt.args.setup(t, root)
			}

			err := a.RemoveAll(
				filepath.Join(root, filepath.FromSlash(tt.args.destRel)),
				tt.args.destRel,
			)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			info, err := os.Stat(filepath.Join(root, filepath.FromSlash(tt.args.destRel)))
			if tt.keepDir {
				its.Nil[error]().Match(err).OrError(t)
				its.EqEq(true).Match(info.IsDir()).OrError(t)

				return
			}

			its.EqEq(true).Match(os.IsNotExist(err)).OrError(t)
		})
	}
}

func TestCopyDir(t *testing.T) {
	t.Parallel()

	type args struct {
		dryRun   bool
		setupSrc func(t *testing.T, root string) string
		destRel  string
	}

	type want struct {
		files    map[string]string
		symlinks []string
	}

	tests := []struct {
		name       string
		args       args
		want       want
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (copies files)",
			args: args{
				setupSrc: func(t *testing.T, root string) string {
					t.Helper()
					src := filepath.Join(root, "src")
					its.Nil[error]().Match(os.MkdirAll(src, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644)).OrFatal(t)

					return src
				},
				destRel: "dest",
			},
			want:       want{files: map[string]string{"a.txt": "a"}},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (dry-run copies nothing)",
			args: args{
				dryRun: true,
				setupSrc: func(t *testing.T, root string) string {
					t.Helper()
					src := filepath.Join(root, "src")
					its.Nil[error]().Match(os.MkdirAll(src, 0o755)).OrFatal(t)

					return src
				},
				destRel: "dest",
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (keeps symlink)",
			args: args{
				setupSrc: func(t *testing.T, root string) string {
					t.Helper()
					src := filepath.Join(root, "src")
					its.Nil[error]().Match(os.MkdirAll(src, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("a.txt", filepath.Join(src, "link"))).OrFatal(t)

					return src
				},
				destRel: "dest",
			},
			want: want{
				files:    map[string]string{"a.txt": "a"},
				symlinks: []string{"link"},
			},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (symlinked dest)",
			args: args{
				setupSrc: func(t *testing.T, root string) string {
					t.Helper()
					src := filepath.Join(root, "src")
					its.Nil[error]().Match(os.MkdirAll(src, 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.MkdirAll(filepath.Join(root, "real"), 0o755)).OrFatal(t)
					its.Nil[error]().Match(os.Symlink("real", filepath.Join(root, "link"))).OrFatal(t)

					return src
				},
				destRel: "link",
			},
			errMatcher: internal.ErrorContaining("symlinked path not allowed"),
		},
		{
			name: "fail (missing src)",
			args: args{
				setupSrc: func(t *testing.T, root string) string {
					t.Helper()

					return filepath.Join(root, "missing")
				},
				destRel: "dest",
			},
			errMatcher: internal.ErrorContaining("missing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, root := newTestApplier(t, tt.args.dryRun)
			src := tt.args.setupSrc(t, root)
			dest := filepath.Join(root, filepath.FromSlash(tt.args.destRel))

			err := a.CopyDir(src, dest, tt.args.destRel)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			if tt.args.dryRun {
				_, err := os.Stat(dest)
				its.EqEq(true).Match(os.IsNotExist(err)).OrError(t)

				return
			}

			for name, wantContent := range tt.want.files {
				got, err := os.ReadFile(filepath.Join(dest, name))
				its.Nil[error]().Match(err).OrError(t)
				its.EqEq(wantContent).Match(string(got)).OrError(t)
			}

			for _, name := range tt.want.symlinks {
				fi, err := os.Lstat(filepath.Join(dest, name))
				its.Nil[error]().Match(err).OrError(t)
				its.EqEq(true).Match(fi.Mode()&fs.ModeSymlink != 0).OrError(t)
			}
		})
	}
}
