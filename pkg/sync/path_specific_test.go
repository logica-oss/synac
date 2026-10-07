package sync_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
	"github.com/logica-oss/synac/pkg/sync"
)

func TestResolvePathSpecific(t *testing.T) {
	t.Parallel()

	type args struct {
		source string
	}

	type want struct {
		srcDir  string
		destDir string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "claude source",
			args: args{source: "claude"},
			want: want{
				srcDir:  ".claude/rules",
				destDir: ".github/instructions",
			},
		},
		{
			name: "github source",
			args: args{source: "github"},
			want: want{
				srcDir:  ".github/instructions",
				destDir: ".claude/rules",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srcDir, destDir := sync.ResolvePathSpecific(tt.args.source)
			its.EqEq(tt.want.srcDir).Match(srcDir).OrError(t)
			its.EqEq(tt.want.destDir).Match(destDir).OrError(t)
		})
	}
}

func TestParseInstruction(t *testing.T) {
	t.Parallel()

	type args struct {
		name    string
		content string
		source  string
		relSrc  string
	}

	tests := []struct {
		name       string
		args       args
		wantName   string
		wantBody   string
		errMatcher its.Matcher[error]
	}{
		{
			name: "success (github to claude)",
			args: args{
				name: "foo.instructions.md",
				content: `---
applyTo: "a, b"
---
hello
`,
				source: "github",
				relSrc: ".github/instructions/foo.instructions.md",
			},
			wantName: "foo.md",
			wantBody: `---
paths:
  - "a"
  - "b"
---

<!-- DO NOT EDIT: Generated from /.github/instructions/foo.instructions.md. Edit /.github/instructions/foo.instructions.md instead. -->

hello
`,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (claude to github)",
			args: args{
				name: "foo.md",
				content: `---
paths:
  - "a"
---
hello
`,
				source: "claude",
				relSrc: ".claude/rules/foo.md",
			},
			wantName: "foo.instructions.md",
			wantBody: `---
applyTo: "a"
---

<!-- DO NOT EDIT: Generated from /.claude/rules/foo.md. Edit /.claude/rules/foo.md instead. -->

hello
`,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (claude to github without globs)",
			args: args{
				name: "foo.md",
				content: `hello
`,
				source: "claude",
				relSrc: ".claude/rules/foo.md",
			},
			wantName: "foo.instructions.md",
			wantBody: `<!-- DO NOT EDIT: Generated from /.claude/rules/foo.md. Edit /.claude/rules/foo.md instead. -->

hello
`,
			errMatcher: its.Nil[error](),
		},
		{
			name: "success (github to claude without globs)",
			args: args{
				name: "foo.instructions.md",
				content: `hello
`,
				source: "github",
				relSrc: ".github/instructions/foo.instructions.md",
			},
			wantName: "foo.md",
			wantBody: `<!-- DO NOT EDIT: Generated from /.github/instructions/foo.instructions.md. Edit /.github/instructions/foo.instructions.md instead. -->

hello
`,
			errMatcher: its.Nil[error](),
		},
		{
			name: "fail (bad applyTo)",
			args: args{
				name: "foo.instructions.md",
				content: `---
applyTo: 123
---
hello
`,
				source: "github",
				relSrc: ".github/instructions/foo.instructions.md",
			},
			errMatcher: internal.ErrorContaining("parse foo.instructions.md"),
		},
		{
			name: "fail (bad paths)",
			args: args{
				name: "foo.md",
				content: `---
paths: 123
---
hello
`,
				source: "claude",
				relSrc: ".claude/rules/foo.md",
			},
			errMatcher: internal.ErrorContaining("parse foo.md"),
		},
		{
			name: "fail (unknown source)",
			args: args{
				name:    "foo.md",
				content: "hello\n",
				source:  "off",
				relSrc:  "foo.md",
			},
			errMatcher: internal.ErrorContaining("unknown path-specific source"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotName, gotBody, err := sync.ParseInstruction(tt.args.name, tt.args.content, tt.args.source, tt.args.relSrc)
			tt.errMatcher.Match(err).OrError(t)

			if err != nil {
				return
			}

			its.EqEq(tt.wantName).Match(gotName).OrError(t)
			its.EqEq(tt.wantBody).Match(gotBody).OrError(t)
		})
	}
}
