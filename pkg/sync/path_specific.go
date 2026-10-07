package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/k1LoW/errors"

	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/md"
	"github.com/logica-oss/synac/pkg/safefs"
)

const pathSpecificHeader = "<!-- DO NOT EDIT: Generated from /%s. Edit /%s instead. -->"

func resolvePathSpecific(source string) (string, string) {
	if source == config.SourceClaude {
		return pathClaudeDir, pathGithubDir
	}

	return pathGithubDir, pathClaudeDir
}

func (r *runner) syncPathSpecific(source string) error {
	srcDirRel, destDirRel := resolvePathSpecific(source)
	srcDir := filepath.Join(r.root, filepath.FromSlash(srcDirRel))
	destDir := filepath.Join(r.root, filepath.FromSlash(destDirRel))

	instructions, err := r.listInstructions(srcDirRel, srcDir, source)
	if err != nil {
		return err
	}

	if safefs.SameFile(srcDir, destDir) {
		return errors.WithStack(fmt.Errorf("path-specific source and destination are the same directory: %s", srcDirRel))
	}

	if err := r.applier.RemoveAll(destDirRel); err != nil {
		return err
	}

	if len(instructions) == 0 {
		return nil
	}

	if err := r.applier.MkdirAll(destDirRel, 0o755); err != nil {
		return err
	}

	for _, name := range instructions {
		destRel, err := r.convertInstruction(srcDir, srcDirRel, destDirRel, source, name)
		if err != nil {
			return err
		}

		r.log.Info("synced path-specific", "dest", destRel)
	}

	return nil
}

func (r *runner) listInstructions(srcDirRel, srcDir, source string) ([]string, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			r.log.Info("path-specific source missing, nothing to sync", "dir", srcDirRel)

			return nil, nil
		}

		return nil, errors.WithStack(fmt.Errorf("read source dir %s: %w", srcDirRel, err))
	}

	suffix := ".instructions.md"
	if source == config.SourceClaude {
		suffix = ".md"
	}

	var instructions []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			instructions = append(instructions, e.Name())
		}
	}

	sort.Strings(instructions)

	return instructions, nil
}

func (r *runner) convertInstruction(srcDir, srcDirRel, destDirRel, source, name string) (string, error) {
	srcPath := filepath.Join(srcDir, name)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("read %s: %w", name, err))
	}

	inName, inBody, err := parseInstruction(name, string(data), source, filepath.ToSlash(filepath.Join(srcDirRel, name)))
	if err != nil {
		return "", err
	}

	destRel := filepath.ToSlash(filepath.Join(destDirRel, inName))

	if err := r.applier.WriteFile(destRel, []byte(inBody), 0o644); err != nil {
		return "", err
	}

	return destRel, nil
}

func parseInstruction(name, content, source, relSrc string) (string, string, error) {
	switch source {
	case config.SourceClaude:
		globs, err := md.ParsePaths(content)
		if err != nil {
			return "", "", errors.WithStack(fmt.Errorf("parse %s: %w", name, err))
		}

		frontmatter := ""
		if len(globs) > 0 {
			frontmatter = md.BuildApplyToFrontmatter(globs)
		}

		return strings.TrimSuffix(name, ".md") + ".instructions.md",
			buildInstructionBody(frontmatter,
				md.StripGeneratedHeader(md.Body(content)), relSrc),
			nil

	case config.SourceGithub:
		globs, err := md.ParseApplyTo(content)
		if err != nil {
			return "", "", errors.WithStack(fmt.Errorf("parse %s: %w", name, err))
		}

		frontmatter := ""
		if len(globs) > 0 {
			frontmatter = md.BuildPathsFrontmatter(globs)
		}

		return strings.TrimSuffix(name, ".instructions.md") + ".md",
			buildInstructionBody(frontmatter,
				md.StripGeneratedHeader(md.Body(content)), relSrc),
			nil

	default:
		return "", "", errors.WithStack(fmt.Errorf("unknown path-specific source %q", source))
	}
}

func buildInstructionBody(frontmatter, body, relSrc string) string {
	var builder strings.Builder

	if frontmatter != "" {
		builder.WriteString(frontmatter)
		builder.WriteString("\n")
	}
	fmt.Fprintf(&builder, pathSpecificHeader+"\n\n", relSrc, relSrc)
	builder.WriteString(body)

	return builder.String()
}
