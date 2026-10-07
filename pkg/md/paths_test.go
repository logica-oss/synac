package md_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/md"
)

//nolint:dupl // table shape mirrors TestBuildApplyToFrontmatter by design
func TestBuildPathsFrontmatter(t *testing.T) {
	t.Parallel()

	type args struct {
		globs []string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{globs: nil},
			want: `---
paths:
---
`,
		},
		{
			name: "single",
			args: args{globs: []string{"a"}},
			want: `---
paths:
  - "a"
---
`,
		},
		{
			name: "multiple",
			args: args{globs: []string{"a", "b"}},
			want: `---
paths:
  - "a"
  - "b"
---
`,
		},
		{
			name: "escaped",
			args: args{globs: []string{`a"b`, `c\d`}},
			want: `---
paths:
  - "a\"b"
  - "c\\d"
---
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(md.BuildPathsFrontmatter(tt.args.globs)).OrError(t)
		})
	}
}

func TestParsePaths(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
	}

	tests := []struct {
		name       string
		args       args
		want       []string
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (no frontmatter)",
			args: args{content: `hello
`},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (no paths key)",
			args: args{content: `---
foo: bar
---
body
`},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (list)",
			args: args{content: `---
paths:
  - "a"
  - "b"
---
body
`},
			want:       []string{"a", "b"},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (string)",
			args: args{content: `---
paths: "a, b"
---
body
`},
			want:       []string{"a", "b"},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (empty list with blanks)",
			args: args{content: `---
paths:
  - ""
  - "a"
---
body
`},
			want:       []string{"a"},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (invalid yaml)",
			args: args{content: `---
foo: [bar
---
body
`},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse paths"),
		},
		{
			name: "fail (unexpected type)",
			args: args{content: `---
paths: 123
---
body
`},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse paths"),
		},
		{
			name: "fail (unexpected item type)",
			args: args{content: `---
paths:
  - 123
---
body
`},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse paths"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := md.ParsePaths(tt.args.content)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.DeepEqual(tt.want).Match(got).OrError(t)
		})
	}
}

func TestExtractPaths(t *testing.T) {
	t.Parallel()

	type args struct {
		v any
	}

	tests := []struct {
		name       string
		args       args
		want       []string
		errMatcher its.Matcher[error]
	}{
		{
			name:       "nil",
			args:       args{v: nil},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name:       "empty string",
			args:       args{v: ""},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name:       "string list",
			args:       args{v: "a, b"},
			want:       []string{"a", "b"},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "empty slice",
			args:       args{v: []any{}},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name:       "slice skips empty",
			args:       args{v: []any{"", "a"}},
			want:       []string{"a"},
			errMatcher: its.Nil[error](),
		},
		{
			name:       "slice unexpected item",
			args:       args{v: []any{123}},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse paths"),
		},
		{
			name:       "unexpected type",
			args:       args{v: 123},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse paths"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := md.ExtractPaths(tt.args.v)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.DeepEqual(tt.want).Match(got).OrError(t)
		})
	}
}
