package md_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/md"
)

func TestBuildApplyToFrontmatter(t *testing.T) {
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
applyTo: ""
---
`,
		},
		{
			name: "single",
			args: args{globs: []string{"a"}},
			want: `---
applyTo: "a"
---
`,
		},
		{
			name: "multiple joined",
			args: args{globs: []string{"a", "b"}},
			want: `---
applyTo: "a, b"
---
`,
		},
		{
			name: "escaped",
			args: args{globs: []string{`a"b`, `c\d`}},
			want: `---
applyTo: "a\"b, c\\d"
---
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(md.BuildApplyToFrontmatter(tt.args.globs)).OrError(t)
		})
	}
}

func TestParseApplyTo(t *testing.T) {
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
			args: args{
				content: `hello
world
`,
			},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (no applyTo key)",
			args: args{
				content: `---
foo: bar
---
body
`,
			},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (simple list)",
			args: args{
				content: `---
applyTo: "a, b"
---
body
`,
			},
			want:       []string{"a", "b"},
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (empty string)",
			args: args{
				content: `---
applyTo: ""
---
body
`,
			},
			want:       nil,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (braces preserved)",
			args: args{
				content: `---
applyTo: "{a,b}, c"
---
body
`,
			},
			want:       []string{"{a,b}", "c"},
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (invalid yaml)",
			args: args{
				content: `---
foo: [bar
---
body
`,
			},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse applyTo"),
		},
		{
			name: "fail (unexpected type int)",
			args: args{
				content: `---
applyTo: 123
---
body
`,
			},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse applyTo"),
		},
		{
			name: "fail (unexpected type list)",
			args: args{
				content: `---
applyTo:
  - "a"
---
body
`,
			},
			want:       nil,
			errMatcher: internal.ErrorContaining("parse applyTo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := md.ParseApplyTo(tt.args.content)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.DeepEqual(tt.want).Match(got).OrError(t)
		})
	}
}
