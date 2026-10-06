package md_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/pkg/md"
)

func TestEscapeYAMLDoubleQuoted(t *testing.T) {
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
			name: "empty",
			args: args{s: ""},
			want: "",
		},
		{
			name: "plain",
			args: args{s: "a/b/*.ts"},
			want: "a/b/*.ts",
		},
		{
			name: "quote escaped",
			args: args{s: `a"b`},
			want: `a\"b`,
		},
		{
			name: "backslash escaped",
			args: args{s: `a\b`},
			want: `a\\b`,
		},
		{
			name: "backslash and quote",
			args: args{s: `a\"b`},
			want: `a\\\"b`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			its.EqEq(tt.want).Match(md.EscapeYAMLDoubleQuoted(tt.args.s)).OrError(t)
		})
	}
}

func TestSplitGlobs(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{s: ""},
			want: nil,
		},
		{
			name: "blank",
			args: args{s: "   "},
			want: nil,
		},
		{
			name: "single",
			args: args{s: "a"},
			want: []string{"a"},
		},
		{
			name: "simple list with spaces",
			args: args{s: "a, b ,c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "empty entries are skipped",
			args: args{s: "a,,b"},
			want: []string{"a", "b"},
		},
		{
			name: "leading and trailing commas",
			args: args{s: ",a,"},
			want: []string{"a"},
		},
		{
			name: "only commas and spaces",
			args: args{s: " , , "},
			want: nil,
		},
		{
			name: "braces preserve commas",
			args: args{s: "{a,b}, c"},
			want: []string{"{a,b}", "c"},
		},
		{
			name: "nested braces",
			args: args{s: "{a,{b,c}}, d"},
			want: []string{"{a,{b,c}}", "d"},
		},
		{
			name: "unmatched closing brace",
			args: args{s: "a},b"},
			want: []string{"a}", "b"},
		},
		{
			name: "brackets preserve commas",
			args: args{s: "[a,b], c"},
			want: []string{"[a,b]", "c"},
		},
		{
			name: "unmatched closing bracket",
			args: args{s: "a],b"},
			want: []string{"a]", "b"},
		},
		{
			name: "brace inside brackets does not change depth",
			args: args{s: "[{], a"},
			want: []string{"[{]", "a"},
		},
		{
			name: "closing brace inside brackets does not change depth",
			args: args{s: "[}], a"},
			want: []string{"[}]", "a"},
		},
		{
			name: "comma inside braces and brackets preserved",
			args: args{s: "{a,b},[c,d],e"},
			want: []string{"{a,b}", "[c,d]", "e"},
		},
		{
			name: "escaped comma",
			args: args{s: `a\,b, c`},
			want: []string{`a\,b`, "c"},
		},
		{
			name: "escaped backslash",
			args: args{s: `a\\, b`},
			want: []string{`a\\`, "b"},
		},
		{
			name: "trailing backslash",
			args: args{s: `a\`},
			want: []string{`a\`},
		},
		{
			name: "escaped brace",
			args: args{s: `\{a, b`},
			want: []string{`\{a`, "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := md.SplitGlobs(tt.args.s)
			its.DeepEqual(tt.want).Match(got).OrError(t)
		})
	}
}
