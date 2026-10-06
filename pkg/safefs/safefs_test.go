package safefs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/logica-oss/synac/pkg/safefs"
	"github.com/youta-t/its"
)

func TestSameFile(t *testing.T) {
	t.Parallel()

	type args struct {
		setup func(t *testing.T, dir string) (string, string)
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same file",
			args: args{
				setup: func(t *testing.T, dir string) (string, string) {
					t.Helper()
					p := filepath.Join(dir, "a.txt")
					its.Nil[error]().Match(os.WriteFile(p, []byte("x"), 0o644)).OrFatal(t)
					return p, p
				},
			},
			want: true,
		},
		{
			name: "different files",
			args: args{
				setup: func(t *testing.T, dir string) (string, string) {
					t.Helper()
					a := filepath.Join(dir, "a.txt")
					b := filepath.Join(dir, "b.txt")
					its.Nil[error]().Match(os.WriteFile(a, []byte("a"), 0o644)).OrFatal(t)
					its.Nil[error]().Match(os.WriteFile(b, []byte("b"), 0o644)).OrFatal(t)
					return a, b
				},
			},
			want: false,
		},
		{
			name: "missing first",
			args: args{
				setup: func(t *testing.T, dir string) (string, string) {
					t.Helper()
					b := filepath.Join(dir, "b.txt")
					its.Nil[error]().Match(os.WriteFile(b, []byte("b"), 0o644)).OrFatal(t)
					return filepath.Join(dir, "missing"), b
				},
			},
			want: false,
		},
		{
			name: "missing second",
			args: args{
				setup: func(t *testing.T, dir string) (string, string) {
					t.Helper()
					a := filepath.Join(dir, "a.txt")
					its.Nil[error]().Match(os.WriteFile(a, []byte("a"), 0o644)).OrFatal(t)
					return a, filepath.Join(dir, "missing")
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			a, b := tt.args.setup(t, dir)

			its.EqEq(tt.want).Match(safefs.SameFile(a, b)).OrError(t)
		})
	}
}
