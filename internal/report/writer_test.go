package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhkvsky/nginxoid/internal/analyzer"
	"github.com/zhkvsky/nginxoid/internal/config"
)

func TestWrite(t *testing.T) {
	files := []string{"logs/access.log", "logs/extra.log"}
	res := analyzer.Result{
		TotalRequests:  3,
		MalformedLines: 1,
		ProcessedLines: 4,
		ResponseCodes: map[int]int{
			200: 2,
			404: 1,
		},
		TopResources: []analyzer.ResourceStat{
			{Resource: "/api/v1/users", Count: 2},
			{Resource: "/health", Count: 1},
		},
		AverageBytes: 256.666,
		MaxBytes:     512,
		P95Bytes:     512,
	}

	tests := []struct {
		name      string
		filename  string
		format    config.Format
		wantErr   bool
		checkFunc func(t *testing.T, outPath string)
	}{
		{
			name:     "valid json report",
			filename: "report.json",
			format:   config.FormatJSON,
			wantErr:  false,
			checkFunc: func(t *testing.T, outPath string) {
				data, err := os.ReadFile(filepath.Clean(outPath))
				require.NoError(t, err)

				var got Report

				err = json.Unmarshal(data, &got)
				require.NoError(t, err)

				assert.Equal(t, files, got.Files)
				assert.Equal(t, 2, got.FilesCount)
				assert.Equal(t, 4, got.ProcessedLinesCount)
				assert.Equal(t, 1, got.MalformedLinesCount)
				assert.Equal(t, 3, got.TotalRequestsCount)
				assert.InDelta(t, 256.67, got.ResponseSizeInBytes.Average, 0.001)
				assert.Equal(t, int64(512), got.ResponseSizeInBytes.Max)
				assert.Equal(t, int64(512), got.ResponseSizeInBytes.P95)
				assert.Equal(t, []resourceItem{
					{Resource: "/api/v1/users", TotalRequestsCount: 2},
					{Resource: "/health", TotalRequestsCount: 1},
				}, got.Resources)
				assert.Equal(t, map[int]int{200: 2, 404: 1}, got.ResponseCodes)
			},
		},
		{
			name:     "valid markdown report",
			filename: "report.md",
			format:   config.FormatMarkdown,
			wantErr:  false,
			checkFunc: func(t *testing.T, outPath string) {
				data, err := os.ReadFile(filepath.Clean(outPath))
				require.NoError(t, err)

				content := string(data)
				assert.Contains(t, content, "# Log Analysis Report")
				assert.Contains(t, content, "- Files count: 2")
				assert.Contains(t, content, "- Processed lines: 4")
				assert.Contains(t, content, "- Malformed lines: 1")
				assert.Contains(t, content, "- Total requests: 3")
				assert.Contains(t, content, "- Average: 256.67")
				assert.Contains(t, content, "| Resource | Requests Count |")
				assert.Contains(t, content, "| /api/v1/users | 2 |")
				assert.Contains(t, content, "| Status Code | Count |")
				assert.Contains(t, content, "| 200 OK | 2 |")
				assert.Contains(t, content, "| 404 Not Found | 1 |")
			},
		},
		{
			name:     "unsupported format returns error",
			filename: "report.xml",
			format:   config.Format("xml"),
			wantErr:  true,
		},
		{
			name:     "non-existent directory returns error",
			filename: filepath.Join("missing_dir", "report.json"),
			format:   config.FormatJSON,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPath := filepath.Join(t.TempDir(), tt.filename)

			err := Write(outPath, tt.format, files, res)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, outPath)
			}
		})
	}
}

func TestWrite_AlreadyExists(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "report.json")
	initialContent := []byte("existing report content")

	err := os.WriteFile(outPath, initialContent, 0o600)
	require.NoError(t, err)

	err = Write(outPath, config.FormatJSON, []string{"access.log"}, analyzer.Result{})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrReportAlreadyExists)

	actualContent, err := os.ReadFile(filepath.Clean(outPath))
	require.NoError(t, err)
	assert.Equal(t, initialContent, actualContent)
}
