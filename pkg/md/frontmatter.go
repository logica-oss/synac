package md

import (
	"strings"
)

const delimiter = "---"

// SplitFrontmatter separates frontmatter from the body.
func SplitFrontmatter(content string) (string, string, bool) {
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
