package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantErr   bool
		targetErr error
	}{
		{
			name:      "valid log file",
			path:      "logs/app.log",
			wantErr:   false,
			targetErr: nil,
		},
		{
			name:      "valid txt file",
			path:      "logs/server.txt",
			wantErr:   false,
			targetErr: nil,
		},
		{
			name:      "valid http url",
			path:      "http://127.0.0.1:8000/access.txt",
			wantErr:   false,
			targetErr: nil,
		},
		{
			name:      "valid https url",
			path:      "https://example.com/log.txt",
			wantErr:   false,
			targetErr: nil,
		},
		{
			name:      "empty path",
			path:      "",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "unsupported extension",
			path:      "data/file.csv",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)

			if tt.wantErr {
				require.Error(t, err)

				if tt.targetErr != nil {
					require.ErrorIs(t, err, tt.targetErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name       string
		rawFormat  string
		wantFormat Format
		wantErr    bool
		targetErr  error
	}{
		{
			name:       "empty format defaults to markdown",
			rawFormat:  "",
			wantFormat: FormatMarkdown,
			wantErr:    false,
		},
		{
			name:       "valid markdown format",
			rawFormat:  "markdown",
			wantFormat: FormatMarkdown,
			wantErr:    false,
		},
		{
			name:       "valid json format",
			rawFormat:  "json",
			wantFormat: FormatJSON,
			wantErr:    false,
		},
		{
			name:      "unsupported format xml",
			rawFormat: "xml",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "unsupported format yaml",
			rawFormat: "yaml",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := validateFormat(tt.rawFormat)

			if tt.wantErr {
				require.Error(t, err)

				if tt.targetErr != nil {
					require.ErrorIs(t, err, tt.targetErr)
				}

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantFormat, format)
		})
	}
}

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		format    Format
		wantErr   bool
		targetErr error
	}{
		{
			name:    "valid markdown output",
			output:  "reports/report.md",
			format:  FormatMarkdown,
			wantErr: false,
		},
		{
			name:    "valid json output",
			output:  "reports/report.json",
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:      "empty output",
			output:    "",
			format:    FormatMarkdown,
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "markdown format with json extension",
			output:    "report.json",
			format:    FormatMarkdown,
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "json format with md extension",
			output:    "report.md",
			format:    FormatJSON,
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "unsupported output extension txt",
			output:    "report.txt",
			format:    FormatMarkdown,
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutput(tt.output, tt.format)

			if tt.wantErr {
				require.Error(t, err)

				if tt.targetErr != nil {
					require.ErrorIs(t, err, tt.targetErr)
				}

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidateDates(t *testing.T) {
	tests := []struct {
		name      string
		rawFrom   string
		rawTo     string
		checkFunc func(t *testing.T, from, to time.Time)
		wantErr   bool
		targetErr error
	}{
		{
			name:    "both dates empty",
			rawFrom: "",
			rawTo:   "",
			checkFunc: func(t *testing.T, from, to time.Time) {
				assert.True(t, from.IsZero())
				assert.True(t, to.IsZero())
			},
			wantErr: false,
		},
		{
			name:    "only from specified",
			rawFrom: "2025-03-01",
			rawTo:   "",
			checkFunc: func(t *testing.T, from, to time.Time) {
				assert.Equal(t, time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), from)
				assert.True(t, to.IsZero())
			},
			wantErr: false,
		},
		{
			name:    "only to specified and shifted to end of day",
			rawFrom: "",
			rawTo:   "2025-03-01",
			checkFunc: func(t *testing.T, from, to time.Time) {
				assert.True(t, from.IsZero())

				expectedTo := time.Date(2025, 3, 1, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
				assert.Equal(t, expectedTo, to)
			},
			wantErr: false,
		},
		{
			name:    "valid range from strictly before to",
			rawFrom: "2025-03-01",
			rawTo:   "2025-03-02",
			checkFunc: func(t *testing.T, from, to time.Time) {
				expectedFrom := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
				expectedTo := time.Date(2025, 3, 2, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

				assert.Equal(t, expectedFrom, from)
				assert.Equal(t, expectedTo, to)
			},
			wantErr: false,
		},
		{
			name:      "equal dates trigger from >= to error",
			rawFrom:   "2025-03-01",
			rawTo:     "2025-03-01",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "from date after to date",
			rawFrom:   "2025-03-05",
			rawTo:     "2025-03-01",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "invalid from date format",
			rawFrom:   "2025-13-40",
			rawTo:     "",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name:      "invalid to date format",
			rawFrom:   "",
			rawTo:     "invalid-date",
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := validateDates(tt.rawFrom, tt.rawTo)

			if tt.wantErr {
				require.Error(t, err)

				if tt.targetErr != nil {
					require.ErrorIs(t, err, tt.targetErr)
				}

				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, from, to)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		rawCfg    RawConfig
		wantErr   bool
		targetErr error
		check     func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid full configuration with default markdown",
			rawCfg: RawConfig{
				Path:   "logs/*.log",
				Format: "",
				Output: "report.md",
				From:   "2025-03-01",
				To:     "2025-03-10",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "logs/*.log", cfg.Path)
				assert.Equal(t, FormatMarkdown, cfg.Format)
				assert.Equal(t, "report.md", cfg.Output)
				assert.Equal(t, 2025, cfg.From.Year())
				assert.Equal(t, 23, cfg.To.Hour())
			},
		},
		{
			name: "valid json config via url",
			rawCfg: RawConfig{
				Path:   "http://127.0.0.1:8000/access.txt",
				Format: "json",
				Output: "out.json",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "http://127.0.0.1:8000/access.txt", cfg.Path)
				assert.Equal(t, FormatJSON, cfg.Format)
				assert.Equal(t, "out.json", cfg.Output)
				assert.True(t, cfg.From.IsZero())
				assert.True(t, cfg.To.IsZero())
			},
		},
		{
			name: "fails on invalid path",
			rawCfg: RawConfig{
				Path:   "logs/app.csv",
				Format: "markdown",
				Output: "report.md",
			},
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
		{
			name: "fails on format and output mismatch",
			rawCfg: RawConfig{
				Path:   "logs/app.log",
				Format: "json",
				Output: "report.md",
			},
			wantErr:   true,
			targetErr: ErrInvalidUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Validate(tt.rawCfg)

			if tt.wantErr {
				require.Error(t, err)

				if tt.targetErr != nil {
					require.ErrorIs(t, err, tt.targetErr)
				}

				assert.Nil(t, cfg)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	t.Run("parses long flags correctly", func(t *testing.T) {
		args := []string{
			"--path", "logs/*.log",
			"--format", "json",
			"--output", "report.json",
			"--from", "2025-03-01",
			"--to", "2025-03-02",
		}

		raw, err := ParseFlags(args)
		require.NoError(t, err)
		assert.Equal(t, "logs/*.log", raw.Path)
		assert.Equal(t, "json", raw.Format)
		assert.Equal(t, "report.json", raw.Output)
		assert.Equal(t, "2025-03-01", raw.From)
		assert.Equal(t, "2025-03-02", raw.To)
	})

	t.Run("parses short flags correctly", func(t *testing.T) {
		args := []string{
			"-p", "logs/app.log",
			"-f", "markdown",
			"-o", "report.md",
		}

		raw, err := ParseFlags(args)
		require.NoError(t, err)
		assert.Equal(t, "logs/app.log", raw.Path)
		assert.Equal(t, "markdown", raw.Format)
		assert.Equal(t, "report.md", raw.Output)
	})

	t.Run("returns ErrInvalidUsage on unknown flag", func(t *testing.T) {
		args := []string{"--invalid-flag", "value"}

		_, err := ParseFlags(args)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidUsage)
	})
}
