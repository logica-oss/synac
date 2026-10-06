package md_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/pkg/md"
)

func TestBody(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "strips frontmatter and leading blanks",
			args: args{content: `---
foo: bar
---


hello
`},
			want: `hello
`,
		},
		{
			name: "no frontmatter trims leading blanks",
			args: args{content: `

hello
`},
			want: `hello
`,
		},
		{
			name: "empty",
			args: args{content: ""},
			want: "",
		},
		{
			name: "only blanks",
			args: args{content: `

`},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(md.Body(tt.args.content)).OrError(t)
		})
	}
}

func TestStripGeneratedHeader(t *testing.T) {
	t.Parallel()

	type args struct {
		body string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "removes header",
			args: args{body: `<!-- DO NOT EDIT: Generated -->
hello
`},
			want: `hello
`,
		},
		{
			name: "removes header with leading blanks",
			args: args{body: `

<!-- DO NOT EDIT: Generated -->

hello
`},
			want: `hello
`,
		},
		{
			name: "removes header with leading spaces",
			args: args{body: `  <!-- DO NOT EDIT: foo -->
hello
`},
			want: `hello
`,
		},
		{
			name: "keeps body without header",
			args: args{body: `hello
world
`},
			want: `hello
world
`,
		},
		{
			name: "keeps similar comment",
			args: args{body: `<!-- EDIT: foo -->
hello
`},
			want: `<!-- EDIT: foo -->
hello
`,
		},
		{
			name: "header only",
			args: args{body: `<!-- DO NOT EDIT: Generated -->
`},
			want: "",
		},
		{
			name: "empty",
			args: args{body: ""},
			want: "",
		},
		{
			name: "only blanks",
			args: args{body: `

`},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(md.StripGeneratedHeader(tt.args.body)).OrError(t)
		})
	}
}

func TestTrimLeadingBlankLines(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no leading blanks",
			args: args{s: `a
b
`},
			want: `a
b
`,
		},
		{
			name: "leading blanks removed",
			args: args{s: `
 
	a
b
`},
			want: `	a
b
`,
		},
		{
			name: "empty",
			args: args{s: ""},
			want: "",
		},
		{
			name: "only blanks",
			args: args{s: `
  
`},
			want: "",
		},
		{
			name: "blank lines inside kept",
			args: args{s: `
a

b
`},
			want: `a

b
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(md.TrimLeadingBlankLines(tt.args.s)).OrError(t)
		})
	}
}
