// Package md parses frontmatter for agent configs.
package md

import (
	"strings"
)

func escapeYAMLDoubleQuoted(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")

	return strings.ReplaceAll(s, "\"", "\\\"")
}

func splitGlobs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	var (
		out []string
		cur strings.Builder
	)
	braceDepth := 0
	bracketDepth := 0

	flush := func() {
		if trimmed := strings.TrimSpace(cur.String()); trimmed != "" {
			out = append(out, trimmed)
		}

		cur.Reset()
	}

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == '\\' && i+1 < len(s) {
			cur.WriteByte(c)
			i++
			cur.WriteByte(s[i])

			continue
		}

		switch c {
		case '{':
			if bracketDepth == 0 {
				braceDepth++
			}
			cur.WriteByte(c)

		case '}':
			if bracketDepth == 0 && braceDepth > 0 {
				braceDepth--
			}
			cur.WriteByte(c)

		case '[':
			bracketDepth++
			cur.WriteByte(c)

		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			cur.WriteByte(c)

		case ',':
			if braceDepth == 0 && bracketDepth == 0 {
				flush()
			} else {
				cur.WriteByte(c)
			}

		default:
			cur.WriteByte(c)
		}
	}

	flush()

	return out
}
