// Command tidy runs go mod tidy for every module in the repository and pins
// each go directive to the lowest version that mingo computes.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/k1LoW/errors"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "tidy: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	modules, err := findModules(root)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return errNoModule
	}

	for _, m := range modules {
		if err := tidyModule(ctx, m); err != nil {
			return errors.WithStack(fmt.Errorf("%s: %w", m.dir, err))
		}
	}

	return nil
}
