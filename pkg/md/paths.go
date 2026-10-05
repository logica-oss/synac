package md

import (
	"fmt"
	"strings"

	"github.com/k1LoW/errors"
	"go.yaml.in/yaml/v3"
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
func ParsePaths(content string) ([]string, error) {
	fm, _, ok := SplitFrontmatter(content)
	if !ok {
		return nil, nil
	}

	var matter struct {
		Paths any `yaml:"paths"`
	}
	if err := yaml.Unmarshal([]byte(fm), &matter); err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse paths: %w", err))
	}

	return extractPaths(matter.Paths)
}

func extractPaths(v any) ([]string, error) {
	switch v := v.(type) {
	case nil:
		return nil, nil

	case string:
		if v == "" {
			return nil, nil
		}
		return []string{v}, nil

	case []any:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out, nil

	default:
		return nil, errors.WithStack(fmt.Errorf("parse paths: unexpected type %T", v))
	}
}
