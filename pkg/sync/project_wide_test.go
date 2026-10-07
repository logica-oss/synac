package sync_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/pkg/sync"
)

func TestResolveProjectWide(t *testing.T) {
	t.Parallel()

	type args struct {
		source string
	}

	type want struct {
		src  string
		dest string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "agents source",
			args: args{source: "agents"},
			want: want{
				src:  "AGENTS.md",
				dest: ".github/copilot-instructions.md",
			},
		},
		{
			name: "github source",
			args: args{source: "github"},
			want: want{
				src:  ".github/copilot-instructions.md",
				dest: "AGENTS.md",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src, dest := sync.ResolveProjectWide(tt.args.source)
			its.DeepEqual(tt.want).Match(want{src: src, dest: dest}).OrError(t)
		})
	}
}
