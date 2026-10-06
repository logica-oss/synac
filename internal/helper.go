// Package internal provides shared helpers for tests.
package internal

import "github.com/youta-t/its"

// ErrorContaining matches any error whose message contains want.
func ErrorContaining(want string) its.Matcher[error] {
	return its.Property(
		"error message",
		func(err error) string {
			if err == nil {
				return "<nil>"
			}
			return err.Error()
		},
		its.StringContaining(want),
	)
}
