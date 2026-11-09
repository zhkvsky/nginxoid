// Package config handles parsing, typing, and validating application startup flags.
package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Format represents the supported output report format.
type Format string

// Supported report formats.
const (
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// RawConfig holds unvalidated raw string CLI arguments.
type RawConfig struct {
	Path   string
	Format string
	Output string
	From   string
	To     string
}

// Config represents validated application parameters parsed into target types.
type Config struct {
	Path   string
	Format Format
	Output string
	From   time.Time
	To     time.Time
}

// Validate validates raw CLI parameters and constructs a valid Config instance.
func Validate(rawCfg RawConfig) (*Config, error) {
	format, err := validateFormat(rawCfg.Format)
	if err != nil {
		return nil, err
	}

	err = validateOutput(rawCfg.Output, format)
	if err != nil {
		return nil, err
	}

	from, to, err := validateDates(rawCfg.From, rawCfg.To)
	if err != nil {
		return nil, err
	}

	err = validatePath(rawCfg.Path)
	if err != nil {
		return nil, err
	}

	return &Config{
		Path:   rawCfg.Path,
		Format: format,
		Output: rawCfg.Output,
		From:   from,
		To:     to,
	}, nil
}

func validateFormat(rawFormat string) (Format, error) {
	switch rawFormat {
	case "":
		return FormatMarkdown, nil
	case "json", "markdown":
		return Format(rawFormat), nil
	}

	return "", fmt.Errorf("%w: unsupported format: %s", ErrInvalidUsage, rawFormat)
}

func validateOutput(rawOutput string, format Format) error {
	if rawOutput == "" {
		return fmt.Errorf("%w: output path is required", ErrInvalidUsage)
	}

	fp := filepath.Ext(rawOutput)

	var expectedExt string

	switch format {
	case FormatJSON:
		expectedExt = ".json"
	case FormatMarkdown:
		expectedExt = ".md"
	}

	if fp != expectedExt {
		return fmt.Errorf("%w: output file extension %q, expected: %q", ErrInvalidUsage, fp, expectedExt)
	}

	return nil
}

func validateDates(rawFrom, rawTo string) (time.Time, time.Time, error) {
	const dateLayout = "2006-01-02"

	var from, to time.Time

	if rawFrom != "" {
		parsed, err := time.Parse(dateLayout, rawFrom)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid from date: %s", ErrInvalidUsage, rawFrom)
		}

		from = parsed
	}

	if rawTo != "" {
		parsed, err := time.Parse(dateLayout, rawTo)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid to date: %s", ErrInvalidUsage, rawTo)
		}

		to = parsed
	}

	if !from.IsZero() && !to.IsZero() {
		if !from.Before(to) {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: from date must be strictly before to date", ErrInvalidUsage)
		}
	}

	if !to.IsZero() {
		to = to.Add(24*time.Hour - time.Nanosecond)
	}

	return from, to, nil
}

func validatePath(rawPath string) error {
	if rawPath == "" {
		return fmt.Errorf("%w: path is required: %s", ErrInvalidUsage, rawPath)
	}

	isURL := strings.HasPrefix(rawPath, "http://") || strings.HasPrefix(rawPath, "https://")

	if isURL {
		return nil
	}

	ext := filepath.Ext(rawPath)
	if ext != ".log" && ext != ".txt" {
		return fmt.Errorf("%w: extension is not supported: %s", ErrInvalidUsage, rawPath)
	}

	return nil
}
