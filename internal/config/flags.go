package config

import (
	"flag"
	"fmt"
	"io"
)

// ParseFlags parses CLI string arguments into a RawConfig struct.
func ParseFlags(args []string) (RawConfig, error) {
	rawCfg := RawConfig{}

	fs := flag.NewFlagSet("nginxoid", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&rawCfg.Path, "path", "", "path to log file, glob pattern, or URL")
	fs.StringVar(&rawCfg.Path, "p", "", "path to log file, glob pattern or URL (shorthand)")

	fs.StringVar(&rawCfg.Format, "format", "", "report format, json or markdown")
	fs.StringVar(&rawCfg.Format, "f", "", "report format, json or markdown (shorthand)")

	fs.StringVar(&rawCfg.Output, "output", "", "path to output report")
	fs.StringVar(&rawCfg.Output, "o", "", "path to output report (shorthand)")

	fs.StringVar(&rawCfg.From, "from", "", "start date of the period in the format YYYY-MM-DD, inclusive")

	fs.StringVar(&rawCfg.To, "to", "", "end date of the period in the format YYYY-MM-DD, inclusive")

	if err := fs.Parse(args); err != nil {
		return RawConfig{}, fmt.Errorf("%w: %w", ErrInvalidUsage, err)
	}

	return rawCfg, nil
}

// Parse parses command-line arguments and validates the resulting configuration.
func Parse(args []string) (*Config, error) {
	rawCfg, err := ParseFlags(args)
	if err != nil {
		return nil, err
	}

	cfg, err := Validate(rawCfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
