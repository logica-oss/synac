package md

import (
	"fmt"
	"strings"

	"github.com/k1LoW/errors"
	"go.yaml.in/yaml/v3"
)

// ParseApplyTo extracts the applyTo value from frontmatter.
func ParseApplyTo(content string) (string, error) {
	fm, _, ok := SplitFrontmatter(content)
	if !ok {
		return "", nil
	}

	var matter struct {
		ApplyTo string `yaml:"applyTo"`
	}
	if err := yaml.Unmarshal([]byte(fm), &matter); err != nil {
		return "", errors.WithStack(fmt.Errorf("parse applyTo: %w", err))
	}

	return matter.ApplyTo, nil
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
