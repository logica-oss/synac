package md

import (
	"strings"
)

const delimiter = "---"

func SplitFrontmatter(content string) (frontmatter, body string, ok bool) {
	content = strings.ReplaceAll(content, "\r", "")

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != delimiter {
		return "", content, false
	}

	closing := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == delimiter {
			closing = i
			break
		}
	}
	if closing == -1 {
		return "", content, false
	}

	return strings.Join(lines[:closing+1], "\n"), strings.Join(lines[closing+1:], "\n"), true
}

func Body(content string) string {
	_, body, ok := SplitFrontmatter(content)
	if !ok {
		body = content
	}

	return trimLeadingBlankLines(body)
}

func StripGeneratedHeader(body string) string {
	rest := trimLeadingBlankLines(body)

	lines := strings.Split(rest, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "<!-- DO NOT EDIT:") {
		rest = strings.Join(lines[1:], "\n")
	}

	return trimLeadingBlankLines(rest)
}

func trimLeadingBlankLines(s string) string {
	lines := strings.Split(s, "\n")

	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	return strings.Join(lines[i:], "\n")
}
