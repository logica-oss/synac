package md_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/pkg/md"
)

func TestSplitFrontmatter(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
	}

	type want struct {
		frontmatter string
		body        string
		ok          bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success (simple)",
			args: args{content: `---
foo: bar
---
body
`},
			want: want{
				frontmatter: `---
foo: bar
---`,
				body: `body
`,
				ok: true,
			},
		},
		{
			name: "success (empty body)",
			args: args{content: `---
foo: bar
---`},
			want: want{
				frontmatter: `---
foo: bar
---`,
				body: "",
				ok:   true,
			},
		},
		{
			name: "success (crlf normalized)",
			args: args{content: "---\r\nfoo: bar\r\n---\r\nbody\r\n"},
			want: want{
				frontmatter: `---
foo: bar
---`,
				body: `body
`,
				ok: true,
			},
		},
		{
			name: "success (closing at end without body)",
			args: args{content: `---
---`},
			want: want{
				frontmatter: `---
---`,
				body: "",
				ok:   true,
			},
		},
		{
			name: "success (first closing wins)",
			args: args{content: `---
a: b
---
body
---
`},
			want: want{
				frontmatter: `---
a: b
---`,
				body: `body
---
`,
				ok: true,
			},
		},
		{
			name: "no frontmatter (plain body)",
			args: args{content: `hello
world
`},
			want: want{
				frontmatter: "",
				body: `hello
world
`,
				ok: false,
			},
		},
		{
			name: "no frontmatter (empty)",
			args: args{content: ""},
			want: want{
				frontmatter: "",
				body:        "",
				ok:          false,
			},
		},
		{
			name: "no closing delimiter",
			args: args{content: `---
foo: bar
body
`},
			want: want{
				frontmatter: "",
				body: `---
foo: bar
body
`,
				ok: false,
			},
		},
		{
			name: "only opening delimiter",
			args: args{content: "---"},
			want: want{
				frontmatter: "",
				body:        "---",
				ok:          false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			frontmatter, body, ok := md.SplitFrontmatter(tt.args.content)
			its.EqEq(tt.want.frontmatter).Match(frontmatter).OrError(t)
			its.EqEq(tt.want.body).Match(body).OrError(t)
			its.EqEq(tt.want.ok).Match(ok).OrError(t)
		})
	}
}
