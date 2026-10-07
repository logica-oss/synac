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
	for i, glob := range globs {
		escaped[i] = escapeYAMLDoubleQuoted(glob)
	}

	var builder strings.Builder

	builder.WriteString("---\n")
	builder.WriteString("applyTo: \"")
	builder.WriteString(strings.Join(escaped, ", "))
	builder.WriteString("\"\n")
	builder.WriteString("---\n")

	return builder.String()
}

// ParseApplyTo extracts glob patterns from frontmatter.
func ParseApplyTo(content string) ([]string, error) {
	frontmatter, _, ok := SplitFrontmatter(content)
	if !ok {
		return nil, nil
	}

	var matter struct {
		ApplyTo any `yaml:"applyTo"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &matter); err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse applyTo: %w", err))
	}

	switch value := matter.ApplyTo.(type) {
	case nil:
		return nil, nil

	case string:
		return splitGlobs(value), nil

	default:
		return nil, errors.WithStack(fmt.Errorf("parse applyTo: unexpected type %T", value))
	}
}
