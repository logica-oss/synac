package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/k1LoW/errors"
)

func tidyModule(ctx context.Context, m module) error {
	if err := tidy(ctx, m); err != nil {
		return err
	}

	return pinGoDirective(m)
}

// The go command has no library API for tidying, so the command is invoked
// directly and its output streams to the process streams.
func tidy(ctx context.Context, m module) error {
	fmt.Printf("==> go mod tidy (%s)\n", m.dir)

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = m.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.WithStack(cmd.Run())
}
