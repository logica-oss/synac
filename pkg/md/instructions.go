package md

import (
	"strings"
)

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

	end := strings.Index(s[1:], string(quote))
	if end == -1 {
		return "", false
	}

	return s[1 : 1+end], true
}

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

func EscapeYAMLDoubleQuoted(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")

	return strings.ReplaceAll(s, "\"", "\\\"")
}

func BuildPathsFrontmatter(globs []string) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("paths:\n")
	for _, g := range globs {
		b.WriteString("  - \"")
		b.WriteString(EscapeYAMLDoubleQuoted(g))
		b.WriteString("\"\n")
	}
	b.WriteString("---\n")

	return b.String()
}

func ParsePaths(content string) []string {
	fm, _, ok := SplitFrontmatter(content)
	if !ok {
		return nil
	}

	var out []string
	inPaths := false
	for line := range strings.Lines(fm) {
		t := strings.TrimSpace(line)

		if t == delimiter {
			continue
		}

		if t == "paths:" {
			inPaths = true
			continue
		}
		indented := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
		if !indented && !strings.HasPrefix(t, "-") && strings.Contains(t, ":") {
			inPaths = false
			continue
		}
		if !inPaths {
			continue
		}

		if !strings.HasPrefix(t, "-") {
			continue
		}

		t = strings.TrimSpace(strings.Trim(strings.TrimSpace(strings.TrimPrefix(t, "-")), "\"'"))
		if t != "" {
			out = append(out, t)
		}
	}

	return out
}
