package md

import (
	"strings"
)

func escapeYAMLDoubleQuoted(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")

	return strings.ReplaceAll(s, "\"", "\\\"")
}

// BuildPathsFrontmatter builds a paths frontmatter block.
func BuildPathsFrontmatter(globs []string) string {
	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("paths:\n")
	for _, g := range globs {
		b.WriteString("  - \"")
		b.WriteString(escapeYAMLDoubleQuoted(g))
		b.WriteString("\"\n")
	}
	b.WriteString("---\n")

	return b.String()
}

// BuildApplyToFrontmatter builds an applyTo frontmatter block.
func BuildApplyToFrontmatter(globs []string) string {
	escaped := make([]string, len(globs))
	for i, g := range globs {
		escaped[i] = escapeYAMLDoubleQuoted(g)
	}

	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("applyTo: \"")
	b.WriteString(strings.Join(escaped, ", "))
	b.WriteString("\"\n")
	b.WriteString("---\n")

	return b.String()
}

// ParsePaths extracts the paths list from frontmatter.
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

			out = append(out, parseInlinePaths(rest)...)
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

		t = unquoteListItem(strings.TrimPrefix(t, "-"))
		if t != "" {
			out = append(out, t)
		}
	}

	return out
}

func parseInlinePaths(rest string) []string {
	if inner, ok := strings.CutPrefix(rest, "["); ok {
		if end := strings.LastIndex(inner, "]"); end != -1 {
			inner = inner[:end]
		}

		var out []string
		for _, g := range SplitGlobs(inner) {
			if g = unquoteListItem(g); g != "" {
				out = append(out, g)
			}
		}

		return out
	}

	if idx := strings.Index(rest, " #"); idx != -1 {
		rest = strings.TrimSpace(rest[:idx])
	}
	if rest = unquoteListItem(rest); rest == "" || strings.HasPrefix(rest, "#") {
		return nil
	}

	return []string{rest}
}

func unquoteListItem(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
		if v, ok := unquote(s, s[0]); ok {
			return v
		}
	}

	return strings.Trim(s, "\"'")
}
