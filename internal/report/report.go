// Package report handles generating report files.
package report

import (
	"math"

	"github.com/zhkvsky/nginxoid/internal/analyzer"
)

type responseSizeMetrics struct {
	Average float64 `json:"average"`
	Max     int64   `json:"max"`
	P95     int64   `json:"p95"`
}

type resourceItem struct {
	Resource           string `json:"resource"`
	TotalRequestsCount int    `json:"totalRequestsCount"`
}

// Report holds the final report data for serialization into JSON or Markdown.
type Report struct {
	Files               []string            `json:"files"`
	FilesCount          int                 `json:"filesCount"`
	ProcessedLinesCount int                 `json:"processedLinesCount"`
	MalformedLinesCount int                 `json:"malformedLinesCount"`
	TotalRequestsCount  int                 `json:"totalRequestsCount"`
	ResponseSizeInBytes responseSizeMetrics `json:"responseSizeInBytes"`
	Resources           []resourceItem      `json:"resources"`
	ResponseCodes       map[int]int         `json:"responseCodes"`
}

// New constructs a Report from the list of processed files and analysis results.
func New(files []string, res analyzer.Result) Report {
	resources := make([]resourceItem, 0, len(res.TopResources))
	for _, v := range res.TopResources {
		resources = append(resources, resourceItem{
			Resource:           v.Resource,
			TotalRequestsCount: v.Count,
		})
	}

	responseCodes := res.ResponseCodes
	if responseCodes == nil {
		responseCodes = make(map[int]int)
	}

	return Report{
		Files:               files,
		FilesCount:          len(files),
		ProcessedLinesCount: res.ProcessedLines,
		MalformedLinesCount: res.MalformedLines,
		TotalRequestsCount:  res.TotalRequests,
		ResponseSizeInBytes: responseSizeMetrics{
			Average: math.Round(res.AverageBytes*100) / 100,
			Max:     res.MaxBytes,
			P95:     res.P95Bytes,
		},
		Resources:     resources,
		ResponseCodes: responseCodes,
	}
}
