package sync_test

import (
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/pkg/sync"
)

func TestResolveSkills(t *testing.T) {
	t.Parallel()

	type args struct {
		source string
	}

	tests := []struct {
		name    string
		args    args
		wantSrc string
		wantDst string
	}{
		{
			name:    "claude source",
			args:    args{source: "claude"},
			wantSrc: ".claude/skills",
			wantDst: ".agents/skills",
		},
		{
			name:    "agents source",
			args:    args{source: "agents"},
			wantSrc: ".agents/skills",
			wantDst: ".claude/skills",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src, dst := sync.ResolveSkills(tt.args.source)
			its.EqEq(tt.wantSrc).Match(src).OrError(t)
			its.EqEq(tt.wantDst).Match(dst).OrError(t)
		})
	}
}
