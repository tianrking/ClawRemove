package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tianrking/ClawRemove/internal/app"
)

func main() {
	ctx := context.Background()
	exitCode, err := app.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
