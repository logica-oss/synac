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
)

const pathSpecificHeader = "<!-- DO NOT EDIT: Generated from /%s. Edit /%s instead. -->"

func resolvePathSpecific(source string) (srcDir, destDir string) {
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

	if srcInfo, srcErr := os.Stat(srcDir); srcErr == nil {
		if destInfo, destErr := os.Stat(destDir); destErr == nil && os.SameFile(srcInfo, destInfo) {
			return errors.WithStack(fmt.Errorf("path-specific source and destination are the same directory: %s", srcDirRel))
		}
	}

	if err := r.applier.RemoveAll(destDir, destDirRel); err != nil {
		return err
	}

	if len(instructions) == 0 {
		return nil
	}

	if err := r.applier.MkdirAll(destDir, destDirRel, 0o755); err != nil {
		return err
	}

	for _, name := range instructions {
		destRel, err := r.convertInstruction(srcDir, srcDirRel, destDir, destDirRel, source, name)
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

func (r *runner) convertInstruction(srcDir, srcDirRel, destDir, destDirRel, source, name string) (string, error) {
	srcPath := filepath.Join(srcDir, name)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("read %s: %w", name, err))
	}

	in := parseInstruction(name, string(data), source, filepath.ToSlash(filepath.Join(srcDirRel, name)))

	destName := in.Name()
	destPath := filepath.Join(destDir, destName)
	destRel := filepath.ToSlash(filepath.Join(destDirRel, destName))

	if err := r.applier.WriteFile(destPath, destRel, []byte(in.Body()), 0o644); err != nil {
		return "", err
	}

	return destRel, nil
}

func parseInstruction(name, content, source, relSrc string) instruction {
	if source == config.SourceClaude {
		return githubInstruction{
			base:   strings.TrimSuffix(name, ".md"),
			globs:  md.ParsePaths(content),
			body:   md.StripGeneratedHeader(md.Body(content)),
			relSrc: relSrc,
		}
	}

	return claudeRule{
		base:   strings.TrimSuffix(name, ".instructions.md"),
		globs:  md.SplitGlobs(md.ParseApplyTo(content)),
		body:   md.StripGeneratedHeader(md.Body(content)),
		relSrc: relSrc,
	}
}

type instruction interface {
	Name() string
	Body() string
}

type githubInstruction struct {
	base   string
	globs  []string
	body   string
	relSrc string
}

func (g githubInstruction) Name() string {
	return g.base + ".instructions.md"
}

func (g githubInstruction) Body() string {
	var b strings.Builder

	if len(g.globs) > 0 {
		b.WriteString(md.BuildApplyToFrontmatter(g.globs))
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, pathSpecificHeader+"\n\n", g.relSrc, g.relSrc)
	b.WriteString(g.body)

	return b.String()
}

type claudeRule struct {
	base   string
	globs  []string
	body   string
	relSrc string
}

func (c claudeRule) Name() string {
	return c.base + ".md"
}

func (c claudeRule) Body() string {
	var b strings.Builder

	if len(c.globs) > 0 {
		b.WriteString(md.BuildPathsFrontmatter(c.globs))
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, pathSpecificHeader+"\n\n", c.relSrc, c.relSrc)
	b.WriteString(c.body)

	return b.String()
}
