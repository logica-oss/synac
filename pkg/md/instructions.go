package md

import (
	"strconv"
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

	if quote == '"' {
		for i := 1; i < len(s); i++ {
			if s[i] == '\\' && i+1 < len(s) {
				i++

				continue
			}
			if s[i] == '"' {
				if decoded, err := strconv.Unquote(s[:i+1]); err == nil {
					return decoded, true
				}

				return s[1:i], true
			}
		}

		return "", false
	}

	for i := 1; i < len(s); i++ {
		if s[i] != '\'' {
			continue
		}
		if i+1 < len(s) && s[i+1] == '\'' {
			i++

			continue
		}

		return strings.ReplaceAll(s[1:i], "''", "'"), true
	}

	return "", false
}

func unquoteYAMLValue(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' {
		for i := 1; i < len(s); i++ {
			if s[i] == '\\' && i+1 < len(s) {
				i++

				continue
			}
			if s[i] == '"' {
				if decoded, err := strconv.Unquote(s[:i+1]); err == nil {
					return decoded
				}

				break
			}
		}
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'")
	}

	return strings.Trim(s, "\"'")
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

		if after, ok := strings.CutPrefix(t, "paths:"); ok {
			rest := strings.TrimSpace(after)
			if rest == "" || strings.HasPrefix(rest, "#") {
				inPaths = true

				continue
			}
			if inner, ok := strings.CutPrefix(rest, "["); ok {
				if end := strings.LastIndex(inner, "]"); end != -1 {
					inner = inner[:end]
				}
				for _, g := range SplitGlobs(inner) {
					g = unquoteYAMLValue(g)
					if g != "" {
						out = append(out, g)
					}
				}
			} else {
				if idx := strings.Index(rest, " #"); idx != -1 {
					rest = strings.TrimSpace(rest[:idx])
				}
				rest = unquoteYAMLValue(rest)
				if rest != "" && !strings.HasPrefix(rest, "#") {
					out = append(out, rest)
				}
			}

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

		t = unquoteYAMLValue(strings.TrimPrefix(t, "-"))
		if t != "" {
			out = append(out, t)
		}
	}

	return out
}
