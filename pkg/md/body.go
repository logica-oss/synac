package md

import (
	"strings"
)

// Body returns the document body without frontmatter.
func Body(content string) string {
	_, body, ok := SplitFrontmatter(content)
	if !ok {
		body = content
	}

	return trimLeadingBlankLines(body)
}

// StripGeneratedHeader removes a leading generated header.
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
