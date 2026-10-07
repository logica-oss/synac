package md

import (
	"fmt"
	"strings"

	"github.com/k1LoW/errors"
	"go.yaml.in/yaml/v3"
)

// BuildPathsFrontmatter builds a paths frontmatter block.
func BuildPathsFrontmatter(globs []string) string {
	var builder strings.Builder

	builder.WriteString("---\n")
	builder.WriteString("paths:\n")
	for _, glob := range globs {
		builder.WriteString("  - \"")
		builder.WriteString(escapeYAMLDoubleQuoted(glob))
		builder.WriteString("\"\n")
	}
	builder.WriteString("---\n")

	return builder.String()
}

// ParsePaths extracts the paths list from frontmatter.
func ParsePaths(content string) ([]string, error) {
	frontmatter, _, ok := SplitFrontmatter(content)
	if !ok {
		return nil, nil
	}

	var matter struct {
		Paths any `yaml:"paths"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &matter); err != nil {
		return nil, errors.WithStack(fmt.Errorf("parse paths: %w", err))
	}

	return extractPaths(matter.Paths)
}

func extractPaths(value any) ([]string, error) {
	switch value := value.(type) {
	case nil:
		return nil, nil

	case string:
		if value == "" {
			return nil, nil
		}

		return splitGlobs(value), nil

	case []any:
		var out []string
		for _, item := range value {
			str, ok := item.(string)
			if !ok {
				return nil, errors.WithStack(fmt.Errorf("parse paths: unexpected item type %T", item))
			}

			if str != "" {
				out = append(out, str)
			}
		}

		return out, nil

	default:
		return nil, errors.WithStack(fmt.Errorf("parse paths: unexpected type %T", value))
	}
}
