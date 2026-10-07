// Package md parses frontmatter for agent configs.
package md

import (
	"strings"
)

func escapeYAMLDoubleQuoted(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")

	return strings.ReplaceAll(s, "\"", "\\\"")
}

type globSplitter struct {
	out          []string
	cur          strings.Builder
	braceDepth   int
	bracketDepth int
}

func (g *globSplitter) flush() {
	if trimmed := strings.TrimSpace(g.cur.String()); trimmed != "" {
		g.out = append(g.out, trimmed)
	}

	g.cur.Reset()
}

func (g *globSplitter) writeEscaped(input string, i int) int {
	g.cur.WriteByte(input[i])

	i++
	g.cur.WriteByte(input[i])

	return i
}

func (g *globSplitter) writeBrace(char byte) {
	if g.bracketDepth == 0 {
		g.braceDepth++
	}

	g.cur.WriteByte(char)
}

func (g *globSplitter) writeCloseBrace(char byte) {
	if g.bracketDepth == 0 && g.braceDepth > 0 {
		g.braceDepth--
	}

	g.cur.WriteByte(char)
}

func (g *globSplitter) writeOpenBracket(char byte) {
	g.bracketDepth++
	g.cur.WriteByte(char)
}

func (g *globSplitter) writeCloseBracket(char byte) {
	if g.bracketDepth > 0 {
		g.bracketDepth--
	}

	g.cur.WriteByte(char)
}

func (g *globSplitter) writeComma(char byte) {
	if g.braceDepth == 0 && g.bracketDepth == 0 {
		g.flush()
	} else {
		g.cur.WriteByte(char)
	}
}

func splitGlobs(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	splitter := &globSplitter{}

	//nolint:intrange // index is advanced manually for escape sequences
	for i := 0; i < len(input); i++ {
		char := input[i]

		if char == '\\' && i+1 < len(input) {
			i = splitter.writeEscaped(input, i)

			continue
		}

		switch char {
		case '{':
			splitter.writeBrace(char)

		case '}':
			splitter.writeCloseBrace(char)

		case '[':
			splitter.writeOpenBracket(char)

		case ']':
			splitter.writeCloseBracket(char)

		case ',':
			splitter.writeComma(char)

		default:
			splitter.cur.WriteByte(char)
		}
	}

	splitter.flush()

	return splitter.out
}
