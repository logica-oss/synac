package sync

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/md"
	"github.com/logica-oss/synac/pkg/safefs"
)

func resolveSkills(source string) (srcDir, destDir string) {
	if source == config.SourceClaude {
		return skillsClaudeDir, skillsAgentsDir
	}

	return skillsAgentsDir, skillsClaudeDir
}

func (r *runner) syncSkills(source string) error {
	srcDirRel, destDirRel := resolveSkills(source)
	srcDir := filepath.Join(r.root, filepath.FromSlash(srcDirRel))
	destDir := filepath.Join(r.root, filepath.FromSlash(destDirRel))

	if err := r.applier.Within(srcDirRel); err != nil {
		return err
	}

	skills, err := r.listSkills(srcDirRel, srcDir)
	if err != nil {
		return err
	}

	for _, name := range skills {
		if err := safefs.CheckSymlinks(r.root, srcDirRel, filepath.Join(srcDir, name)); err != nil {
			return err
		}
	}

	if safefs.SameFile(srcDir, destDir) {
		return errors.WithStack(fmt.Errorf("skills source and destination are the same directory: %s", srcDirRel))
	}

	if err := r.applier.RemoveAll(destDir, destDirRel); err != nil {
		return err
	}

	if len(skills) == 0 {
		return nil
	}

	if err := r.applier.MkdirAll(destDir, destDirRel, 0o755); err != nil {
		return err
	}

	for _, name := range skills {
		destRel, err := r.copySkill(srcDir, srcDirRel, destDir, destDirRel, name)
		if err != nil {
			return err
		}

		r.log.Info("synced skill", "dest", destRel)
	}

	return nil
}

func (r *runner) listSkills(srcDirRel, srcDir string) ([]string, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			r.log.Info("skills source missing, nothing to sync", "dir", srcDirRel)
			return nil, nil
		}

		return nil, errors.WithStack(fmt.Errorf("read source dir %s: %w", srcDirRel, err))
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}

	sort.Strings(names)

	return names, nil
}

func (r *runner) copySkill(srcDir, srcDirRel, destDir, destDirRel, name string) (string, error) {
	srcSkill := filepath.Join(srcDir, name)
	destSkill := filepath.Join(destDir, name)
	destRel := filepath.ToSlash(filepath.Join(destDirRel, name))

	if err := r.applier.CopyDir(srcSkill, destSkill, destRel); err != nil {
		return "", errors.WithStack(fmt.Errorf("copy skill %s: %w", name, err))
	}

	if err := r.rewriteSkillHeader(srcDirRel, destDirRel, name, srcSkill, destSkill); err != nil {
		return "", err
	}

	return destRel, nil
}

func (r *runner) rewriteSkillHeader(srcDirRel, destDirRel, name, srcSkill, destSkill string) error {
	srcFile := filepath.Join(srcSkill, "SKILL.md")
	data, err := os.ReadFile(srcFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	destFile := filepath.Join(destSkill, "SKILL.md")
	if info, err := os.Lstat(destFile); err == nil && info.Mode()&fs.ModeSymlink != 0 {
		return nil
	}

	content := string(data)

	var b strings.Builder
	if fm, _, ok := md.SplitFrontmatter(content); ok {
		b.WriteString(fm)
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "<!-- DO NOT EDIT: Generated from /%s/%s. Edit /%s/%s instead. -->\n\n", srcDirRel, name, srcDirRel, name)
	b.WriteString(md.StripGeneratedHeader(md.Body(content)))

	destRel := filepath.ToSlash(filepath.Join(destDirRel, name, "SKILL.md"))

	return r.applier.WriteFile(destFile, destRel, []byte(b.String()), 0o644)
}
