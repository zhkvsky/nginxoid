// Package main implements the Incident Radar CLI application.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zhkvsky/nginxoid/internal/config"
	"github.com/zhkvsky/nginxoid/internal/source"
)

func main() {
	ctx := context.Background()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cfg, err := config.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error parsing configuration: %v\\n", err)
		return 2
	}

	src, err := source.Open(ctx, cfg.Path)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error opening path: %v\n", err)
		if errors.Is(err, source.ErrNotFound) {
			return 2
		}

		return 1
	}
	defer func() {
		_ = src.Reader.Close()
	}()

	_, _ = fmt.Fprintf(stdout, "Sources opened: %d\n", len(src.Files))
	for _, file := range src.Files {
		_, _ = fmt.Fprintf(stdout, "- %s\n", file)
	}

	return 0
}
