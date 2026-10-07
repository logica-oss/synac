// Command tidy runs go mod tidy for every module in the repository and pins
// each go directive to the lowest version that mingo computes.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/k1LoW/errors"
)

var errNoModule = errors.New("no go.mod found")

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := run(context.Background()); err != nil {
		slog.Error("tidy failed", "error", err, "stack", errors.StackTraces(err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return errors.WithStack(fmt.Errorf("get working directory: %w", err))
	}

	modules, err := findModules(root)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return errors.WithStack(errNoModule)
	}

	for _, mod := range modules {
		if err := tidy(ctx, mod); err != nil {
			return err
		}

		if err := pinGoDirective(mod); err != nil {
			return err
		}
	}

	return nil
}
