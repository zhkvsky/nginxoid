// Package main implements the Incident Radar CLI application.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zhkvsky/nginxoid/internal/config"
	"github.com/zhkvsky/nginxoid/internal/source"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	rawCfg, err := config.ParseFlags(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	cfg, err := config.Validate(rawCfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error validating config: %v\n", err)
		return 2
	}

	src, err := source.Open(cfg.Path)
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
