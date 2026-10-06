// Package internal provides shared helpers for tests.
package internal

import "github.com/youta-t/its"

// ErrorContaining matches a non-nil error whose message contains want.
func ErrorContaining(want string) its.Matcher[error] {
	return its.All(
		its.Not(its.Nil[error]()),
		its.Property(
			"error message",
			func(err error) string {
				if err == nil {
					return "<nil>"
				}
				return err.Error()
			},
			its.StringContaining(want),
		),
	)
}
