package md

import (
	"fmt"

	"github.com/k1LoW/errors"
	"go.yaml.in/yaml/v3"
)

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
