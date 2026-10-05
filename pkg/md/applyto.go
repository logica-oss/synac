package md

import (
	"strings"
)

// ParseApplyTo extracts the applyTo value from frontmatter.
func ParseApplyTo(content string) string {
	fm, _, ok := SplitFrontmatter(content)
	if !ok {
		return ""
	}

	for line := range strings.Lines(fm) {
		if !strings.HasPrefix(line, "applyTo:") {
			continue
		}

		rest := strings.TrimRight(strings.TrimSpace(strings.TrimPrefix(line, "applyTo:")), "\n")

		if v, ok := unquote(rest, '"'); ok {
			return v
		}
		if v, ok := unquote(rest, '\''); ok {
			return v
		}

		trimmed := strings.Trim(strings.TrimSpace(strings.Trim(rest, "\"'")), "\"'")

		if idx := strings.Index(trimmed, " #"); idx != -1 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}

		return strings.Trim(trimmed, "\"'")
	}

	return ""
}

func unquote(s string, quote byte) (string, bool) {
	if len(s) < 2 || s[0] != quote {
		return "", false
	}

	var b strings.Builder
	for i := 1; i < len(s); i++ {
		c := s[i]
		if quote == '"' && c == '\\' && i+1 < len(s) {
			i++
			b.WriteByte(s[i])
			continue
		}

		if c == quote {
			return b.String(), true
		}

		b.WriteByte(c)
	}

	return "", false
}

// SplitGlobs splits a comma-separated glob list respecting braces.
func SplitGlobs(applyTo string) []string {
	if strings.TrimSpace(applyTo) == "" {
		return nil
	}

	var (
		out []string
		cur strings.Builder
	)
	depth := 0

	flush := func() {
		if trimmed := strings.TrimSpace(cur.String()); trimmed != "" {
			out = append(out, trimmed)
		}

		cur.Reset()
	}

	for i := 0; i < len(applyTo); i++ {
		c := applyTo[i]

		if c == '\\' && i+1 < len(applyTo) {
			cur.WriteByte(c)
			i++
			cur.WriteByte(applyTo[i])

			continue
		}

		switch c {
		case '{':
			depth++
			cur.WriteByte(c)

		case '}':
			if depth > 0 {
				depth--
			}
			cur.WriteByte(c)

		case ',':
			if depth == 0 {
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
