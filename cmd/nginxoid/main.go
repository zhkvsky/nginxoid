// Package main implements the Incident Radar CLI application.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zhkvsky/nginxoid/internal/analyzer"
	"github.com/zhkvsky/nginxoid/internal/config"
	"github.com/zhkvsky/nginxoid/internal/report"
	"github.com/zhkvsky/nginxoid/internal/source"
)

func main() {
	ctx := context.Background()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cfg, err := config.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error parsing configuration: %v\n", err)
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

	stats := analyzer.New()
	if err = stats.Process(src.Reader, cfg.From, cfg.To, stderr); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error processing logs: %v\n", err)
		return 1
	}

	res := stats.Result()

	if err = report.Write(cfg.Output, cfg.Format, src.Files, res); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error writing report: %v\n", err)
		if errors.Is(err, report.ErrReportAlreadyExists) {
			return 2
		}

		return 1
	}

	_, _ = fmt.Fprintf(stdout, "Report saved to %s\n", cfg.Output)

	return 0
}
