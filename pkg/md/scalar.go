// Package md parses frontmatter for agent configs.
package md

import (
	"strings"
)

func repairBareGlobs(fm string) (string, bool) {
	lines := strings.Split(fm, "\n")
	repaired := false
	for i, line := range lines {
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		t := strings.TrimSpace(line)

		if rest, ok := strings.CutPrefix(t, "paths:"); ok {
			if nr, ok := repairPathsValue(strings.TrimSpace(rest)); ok {
				lines[i] = indent + "paths: " + nr
				repaired = true
			}
		} else if rest, ok := strings.CutPrefix(t, "applyTo:"); ok {
			if nr, ok := repairScalarValue(strings.TrimSpace(rest)); ok {
				lines[i] = indent + "applyTo: " + nr
				repaired = true
			}
		} else if rest, ok := strings.CutPrefix(t, "- "); ok {
			if nr, ok := repairScalarValue(strings.TrimSpace(rest)); ok {
				lines[i] = indent + "- " + nr
				repaired = true
			}
		}
	}

	return strings.Join(lines, "\n"), repaired
}

func repairPathsValue(rest string) (string, bool) {
	if rest == "" || strings.HasPrefix(rest, "#") {
		return "", false
	}

	if inner, ok := strings.CutPrefix(rest, "["); ok {
		end := strings.LastIndex(inner, "]")
		if end == -1 {
			return "", false
		}

		tail := inner[end+1:]
		changed := false
		out := make([]string, 0)
		for _, g := range SplitGlobs(inner[:end]) {
			if nq, ok := quoteBare(g); ok {
				g = nq
				changed = true
			}
			out = append(out, g)
		}
		if !changed {
			return "", false
		}

		return "[" + strings.Join(out, ", ") + "]" + tail, true
	}

	return repairScalarValue(rest)
}

func repairScalarValue(rest string) (string, bool) {
	if rest == "" || strings.HasPrefix(rest, "#") {
		return "", false
	}

	return quoteBare(rest)
}

func quoteBare(s string) (string, bool) {
	t := strings.TrimSpace(s)
	if t == "" || t[0] != '*' {
		return "", false
	}

	scalar, comment := splitTrailingComment(t)
	q := "\"" + strings.ReplaceAll(strings.ReplaceAll(scalar, "\\", "\\\\"), "\"", "\\\"") + "\""

	return q + comment, true
}

func splitTrailingComment(s string) (scalar, comment string) {
	if idx := strings.Index(s, " #"); idx != -1 {
		return strings.TrimSpace(s[:idx]), s[idx:]
	}
	if idx := strings.Index(s, "\t#"); idx != -1 {
		return strings.TrimSpace(s[:idx]), s[idx:]
	}

	return s, ""
}
