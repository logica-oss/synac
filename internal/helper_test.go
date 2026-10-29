package internal_test

import (
	"errors"
	"testing"

	"github.com/youta-t/its"

	"github.com/logica-oss/synac/internal"
)

func TestErrorContaining(t *testing.T) {
	t.Parallel()

	type args struct {
		want string
		err  error
	}

	tests := []struct {
		name   string
		args   args
		wantOK bool
	}{
		{
			name:   "match",
			args:   args{want: "boom", err: errors.New("boom happened")},
			wantOK: true,
		},
		{
			name:   "mismatch",
			args:   args{want: "boom", err: errors.New("other failure")},
			wantOK: false,
		},
		{
			name:   "nil error",
			args:   args{want: "boom", err: nil},
			wantOK: false,
		},
		{
			name:   "nil message still fails on nil check",
			args:   args{want: "<nil>", err: nil},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matched := internal.ErrorContaining(tt.args.want).Match(tt.args.err).Ok()
			its.EqEq(tt.wantOK).Match(matched).OrError(t)
		})
	}
}
