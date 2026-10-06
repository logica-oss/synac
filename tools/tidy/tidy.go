package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"

	"github.com/k1LoW/errors"
)

func tidy(ctx context.Context, mod module) error {
	slog.Info("go mod tidy", "dir", mod.dir)

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = mod.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.WithStack(cmd.Run())
}
