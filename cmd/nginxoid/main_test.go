package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	tmpDir := t.TempDir()

	binName := "incident-radar"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	binPath := filepath.Join(tmpDir, binName)

	buildCmd := exec.CommandContext(t.Context(), "go", "build", "-o", binPath, ".") //nolint:gosec // test builds local binary

	buildCmd.Env = append(os.Environ(), "GOROOT=")
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build binary: %s", string(buildOut))

	logPath := filepath.Join(tmpDir, "access.log")
	logContent := "93.180.71.3 - - [17/May/2015:08:05:32 +0000] \"GET /downloads/product_1 HTTP/1.1\" 200 512 \"-\" \"Agent\"\n" +
		"broken malformed line\n"

	err = os.WriteFile(logPath, []byte(logContent), 0o600)
	require.NoError(t, err)

	runCLI := func(args ...string) (string, int) {
		var stdout, stderr bytes.Buffer

		cmd := exec.CommandContext(t.Context(), binPath, args...) //nolint:gosec // test runs compiled binary
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		runErr := cmd.Run()
		if runErr == nil {
			return stderr.String(), 0
		}

		var exitErr *exec.ExitError
		require.ErrorAs(t, runErr, &exitErr)

		return stderr.String(), exitErr.ExitCode()
	}

	t.Run("success json report with exit code 0", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "report.json")

		stderr, code := runCLI("--path", logPath, "--format", "json", "--output", outPath)
		assert.Equal(t, 0, code)
		assert.NotEmpty(t, stderr)

		data, readErr := os.ReadFile(filepath.Clean(outPath))
		require.NoError(t, readErr)

		var parsed struct {
			FilesCount          int `json:"filesCount"`
			ProcessedLinesCount int `json:"processedLinesCount"`
			MalformedLinesCount int `json:"malformedLinesCount"`
			TotalRequestsCount  int `json:"totalRequestsCount"`
		}

		require.NoError(t, json.Unmarshal(data, &parsed))
		assert.Equal(t, 1, parsed.FilesCount)
		assert.Equal(t, 2, parsed.ProcessedLinesCount)
		assert.Equal(t, 1, parsed.MalformedLinesCount)
		assert.Equal(t, 1, parsed.TotalRequestsCount)
	})

	t.Run("missing path returns exit code 2", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "report.json")

		stderr, code := runCLI("--format", "json", "--output", outPath)
		assert.Equal(t, 2, code)
		assert.NotEmpty(t, stderr)
	})

	t.Run("invalid date range returns exit code 2", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "report.json")

		stderr, code := runCLI("--path", logPath, "--output", outPath, "--from", "2025-05-18", "--to", "2025-05-17")
		assert.Equal(t, 2, code)
		assert.NotEmpty(t, stderr)
	})

	t.Run("missing source file returns exit code 2", func(t *testing.T) {
		missingLog := filepath.Join(t.TempDir(), "missing.log")
		outPath := filepath.Join(t.TempDir(), "report.json")

		stderr, code := runCLI("--path", missingLog, "--output", outPath)
		assert.Equal(t, 2, code)
		assert.NotEmpty(t, stderr)
	})

	t.Run("existing output file returns exit code 2 without overwrite", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "report.json")
		initial := []byte("do not overwrite")

		require.NoError(t, os.WriteFile(outPath, initial, 0o600))

		stderr, code := runCLI("--path", logPath, "--format", "json", "--output", outPath)
		assert.Equal(t, 2, code)
		assert.NotEmpty(t, stderr)

		actual, readErr := os.ReadFile(filepath.Clean(outPath))
		require.NoError(t, readErr)
		assert.Equal(t, initial, actual)
	})

	t.Run("unwritable output directory returns exit code 1", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "missing_dir", "report.json")

		stderr, code := runCLI("--path", logPath, "--format", "json", "--output", outPath)
		assert.Equal(t, 1, code)
		assert.NotEmpty(t, stderr)
	})
}
