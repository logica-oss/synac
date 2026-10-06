package md

import (
	"fmt"
	"strings"

	"github.com/k1LoW/errors"
	"go.yaml.in/yaml/v3"
)

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

// ParseApplyTo extracts glob patterns from frontmatter.
func ParseApplyTo(content string) ([]string, error) {
	fm, _, ok := SplitFrontmatter(content)
	if !ok {
		return nil, nil
	}

	var matter struct {
		ApplyTo any `yaml:"applyTo"`
	}
	if err := yaml.Unmarshal([]byte(fm), &matter); err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse applyTo: %w", err))
	}

	switch v := matter.ApplyTo.(type) {
	case nil:
		return nil, nil

	case string:
		return splitGlobs(v), nil

	default:
		return nil, errors.WithStack(fmt.Errorf("parse applyTo: unexpected type %T", v))
	}
}
