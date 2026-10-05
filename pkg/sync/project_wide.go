package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
	"github.com/logica-oss/synac/pkg/config"
	"github.com/logica-oss/synac/pkg/md"
)

const projectWideHeader = "<!-- DO NOT EDIT: Generated mirror of /%s. Edit /%s instead. -->"

func resolveProjectWide(source string) (src, dest string) {
	if source == config.SourceAgents {
		return projectWideRootFile, projectWideGithubFile
	}

	return projectWideGithubFile, projectWideRootFile
}

func (r *runner) syncProjectWide(source string) error {
	srcRel, destRel := resolveProjectWide(source)
	src := filepath.Join(r.root, filepath.FromSlash(srcRel))
	dest := filepath.Join(r.root, filepath.FromSlash(destRel))

	data, err := os.ReadFile(src)
	if err != nil {
		return errors.WithStack(fmt.Errorf("project-wide source not found: %s: %w", srcRel, err))
	}
	content := string(data)

	body := md.StripGeneratedHeader(md.Body(content))
	out := fmt.Sprintf(projectWideHeader+"\n\n", srcRel, srcRel) + body

	if err := r.applier.WriteFile(dest, destRel, []byte(out), 0o644); err != nil {
		return err
	}

	r.log.Info("synced project-wide", "src", srcRel, "dest", destRel)

	return nil
}
