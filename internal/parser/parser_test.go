package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLine(t *testing.T) {
	t.Run("success standard nginx line", func(t *testing.T) {
		line := `93.180.71.3 - - [17/May/2015:08:05:32 +0000] "GET /downloads/product_1 HTTP/1.1" 304 0 "-" "Debian APT-HTTP/1.3"`
		rec, err := ParseLine(line)

		require.NoError(t, err)
		require.NotNil(t, rec)

		assert.Equal(t, "93.180.71.3", rec.RemoteAddr)
		assert.Equal(t, "-", rec.RemoteUser)
		assert.Equal(t, "GET", rec.Method)
		assert.Equal(t, "/downloads/product_1", rec.Resource)
		assert.Equal(t, "HTTP/1.1", rec.Protocol)
		assert.Equal(t, "-", rec.HTTPReferer)
		assert.Equal(t, "Debian APT-HTTP/1.3", rec.HTTPUserAgent)

		assert.Equal(t, 304, rec.Status)
		assert.Equal(t, int64(0), rec.BodyBytesSent)

		expectedTime := time.Date(2015, time.May, 17, 8, 5, 32, 0, time.UTC)
		assert.True(t, expectedTime.Equal(rec.TimeLocal))
	})

	t.Run("malformed lines return ErrMalformedLine", func(t *testing.T) {
		tests := []struct {
			name string
			line string
		}{
			{
				name: "empty string",
				line: "",
			},
			{
				name: "missing request quotes",
				line: `93.180.71.3 - - [17/May/2015:08:05:32 +0000] GET /index.html HTTP/1.1 200 100 "-" "-"`,
			},
			{
				name: "invalid status code (not a number)",
				line: `93.180.71.3 - - [17/May/2015:08:05:32 +0000] "GET /test HTTP/1.1" OK 100 "-" "-"`,
			},
			{
				name: "invalid time format",
				line: `93.180.71.3 - - [2015-05-17 08:05:32] "GET /test HTTP/1.1" 200 100 "-" "-"`,
			},
			{
				name: "incomplete request without protocol",
				line: `93.180.71.3 - - [17/May/2015:08:05:32 +0000] "GET /test" 200 100 "-" "-"`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rec, err := ParseLine(tt.line)

				require.Error(t, err)
				require.ErrorIs(t, err, ErrMalformedLine)
				assert.Nil(t, rec)
			})
		}
	})
}
