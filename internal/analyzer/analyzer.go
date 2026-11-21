package analyzer

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"time"

	"github.com/zhkvsky/nginxoid/internal/parser"
)

const (
	maxTopResources    = 10
	initialScanBufSize = 64 * 1024
	maxScanLineSize    = 1024 * 1024
)

// Analyzer accumulates statistics from processed log records.
type Analyzer struct {
	requestCount   int
	malformedCount int
	statusCodes    map[int]int
	resources      map[string]int
	totalBytes     int64
	maxBytes       int64
	responseSizes  []int64
}

// ResourceStat holds request count statistics for a single resource.
type ResourceStat struct {
	Resource string
	Count    int
}

// Result contains the final aggregated metrics after log processing.
type Result struct {
	TotalRequests  int
	MalformedLines int
	ResponseCodes  map[int]int
	TopResources   []ResourceStat
	AverageBytes   float64
	MaxBytes       int64
	P95Bytes       int64
}

// New creates and initializes a new Analyzer instance.
func New() *Analyzer {
	return &Analyzer{
		statusCodes: make(map[int]int),
		resources:   make(map[string]int),
	}
}

// AddMalformed increments the counter for skipped malformed log lines.
func (a *Analyzer) AddMalformed() {
	a.malformedCount++
}

// Add updates the aggregated metrics with the given log record.
func (a *Analyzer) Add(rec *parser.LogRecord) {
	a.requestCount++

	a.statusCodes[rec.Status]++

	a.resources[rec.Resource]++

	a.totalBytes += rec.BodyBytesSent
	if rec.BodyBytesSent > a.maxBytes {
		a.maxBytes = rec.BodyBytesSent
	}

	a.responseSizes = append(a.responseSizes, rec.BodyBytesSent)
}

// Process reads log lines from stream r, filters them by the [from, to] date range, and aggregates statistics.
func (a *Analyzer) Process(r io.Reader, from, to time.Time, warnOut io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, initialScanBufSize), maxScanLineSize)

	for scanner.Scan() {
		line := scanner.Text()

		rec, err := parser.ParseLine(line)
		if err != nil {
			if warnOut != nil {
				_, _ = fmt.Fprintf(warnOut, "WARN: skipped malformed log line: %v\n", err)
			}

			a.AddMalformed()

			continue
		}

		if !MatchesDateRange(rec.TimeLocal, from, to) {
			continue
		}

		a.Add(rec)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read log stream: %w", err)
	}

	return nil
}

// Result computes the final metrics (average, p95, top 10 resources) and returns the summary.
func (a *Analyzer) Result() Result {
	var (
		avg float64
		p95 int64
	)

	if a.requestCount > 0 {
		avg = float64(a.totalBytes) / float64(a.requestCount)

		slices.Sort(a.responseSizes)
		idx := int(math.Ceil(0.95*float64(len(a.responseSizes)))) - 1
		p95 = a.responseSizes[idx]
	}

	topResources := make([]ResourceStat, 0, len(a.resources))
	for res, count := range a.resources {
		topResources = append(topResources, ResourceStat{
			Resource: res,
			Count:    count,
		})
	}

	slices.SortFunc(topResources, func(x, y ResourceStat) int {
		return cmp.Or(cmp.Compare(y.Count, x.Count), cmp.Compare(x.Resource, y.Resource))
	})

	if len(topResources) > maxTopResources {
		topResources = topResources[:maxTopResources]
	}

	return Result{
		TotalRequests:  a.requestCount,
		MalformedLines: a.malformedCount,
		ResponseCodes:  maps.Clone(a.statusCodes),
		TopResources:   topResources,
		AverageBytes:   avg,
		MaxBytes:       a.maxBytes,
		P95Bytes:       p95,
	}
}
