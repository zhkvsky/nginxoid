package analyzer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhkvsky/nginxoid/internal/parser"
)

func TestMatchesDateRange(t *testing.T) {
	baseTime := time.Date(2025, time.March, 5, 12, 0, 0, 0, time.UTC)
	fromTime := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	toTime := time.Date(2025, time.March, 10, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		recordTime time.Time
		from       time.Time
		to         time.Time
		want       bool
	}{
		{
			name:       "no date bounds",
			recordTime: baseTime,
			want:       true,
		},
		{
			name:       "only from bound - inside",
			recordTime: baseTime,
			from:       fromTime,
			want:       true,
		},
		{
			name:       "only from bound - exact inclusive boundary",
			recordTime: fromTime,
			from:       fromTime,
			want:       true,
		},
		{
			name:       "only from bound - before",
			recordTime: fromTime.Add(-time.Second),
			from:       fromTime,
			want:       false,
		},
		{
			name:       "only to bound - inside",
			recordTime: baseTime,
			to:         toTime,
			want:       true,
		},
		{
			name:       "only to bound - exact inclusive boundary",
			recordTime: toTime,
			to:         toTime,
			want:       true,
		},
		{
			name:       "only to bound - after",
			recordTime: toTime.Add(time.Second),
			to:         toTime,
			want:       false,
		},
		{
			name:       "both bounds - inside",
			recordTime: baseTime,
			from:       fromTime,
			to:         toTime,
			want:       true,
		},
		{
			name:       "both bounds - out of range before",
			recordTime: fromTime.Add(-time.Hour),
			from:       fromTime,
			to:         toTime,
			want:       false,
		},
		{
			name:       "both bounds - out of range after",
			recordTime: toTime.Add(time.Hour),
			from:       fromTime,
			to:         toTime,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesDateRange(tt.recordTime, tt.from, tt.to)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnalyzer_MetricsAndControlDataset(t *testing.T) {
	a := New()

	records := []*parser.LogRecord{
		{Resource: "/a", Status: 200, BodyBytesSent: 300},
		{Resource: "/a", Status: 200, BodyBytesSent: 100},
		{Resource: "/b", Status: 404, BodyBytesSent: 500},
		{Resource: "/a", Status: 200, BodyBytesSent: 200},
		{Resource: "/b", Status: 404, BodyBytesSent: 400},
	}

	for _, rec := range records {
		a.Add(rec)
	}

	a.AddMalformed()
	a.AddMalformed()

	res := a.Result()

	assert.Equal(t, 5, res.TotalRequests)
	assert.Equal(t, 2, res.MalformedLines)
	assert.InDelta(t, 300.0, res.AverageBytes, 1e-9)
	assert.Equal(t, int64(500), res.MaxBytes)
	assert.Equal(t, int64(500), res.P95Bytes)
	assert.Equal(t, map[int]int{200: 3, 404: 2}, res.ResponseCodes)
}

func TestAnalyzer_TopResourcesSortingAndLimit(t *testing.T) {
	a := New()

	tied := []string{"/api/c", "/api/a", "/api/b"}
	for _, r := range tied {
		for range 3 {
			a.Add(&parser.LogRecord{Resource: r, Status: 200, BodyBytesSent: 100})
		}
	}

	for i := 1; i <= 9; i++ {
		a.Add(&parser.LogRecord{
			Resource:      fmt.Sprintf("/page/%d", i),
			Status:        200,
			BodyBytesSent: 100,
		})
	}

	res := a.Result()

	require.Len(t, res.TopResources, 10)
	assert.Equal(t, ResourceStat{Resource: "/api/a", Count: 3}, res.TopResources[0])
	assert.Equal(t, ResourceStat{Resource: "/api/b", Count: 3}, res.TopResources[1])
	assert.Equal(t, ResourceStat{Resource: "/api/c", Count: 3}, res.TopResources[2])
	assert.Equal(t, ResourceStat{Resource: "/page/1", Count: 1}, res.TopResources[3])
	assert.Equal(t, ResourceStat{Resource: "/page/7", Count: 1}, res.TopResources[9])
}

func TestAnalyzer_Empty(t *testing.T) {
	a := New()
	res := a.Result()

	assert.Equal(t, 0, res.TotalRequests)
	assert.Equal(t, 0, res.MalformedLines)
	assert.InDelta(t, 0.0, res.AverageBytes, 1e-9)
	assert.Equal(t, int64(0), res.MaxBytes)
	assert.Equal(t, int64(0), res.P95Bytes)
	assert.Empty(t, res.TopResources)
}
