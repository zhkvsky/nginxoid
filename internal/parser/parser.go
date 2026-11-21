// Package parser handles parsing of NGINX log lines.
package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/zhkvsky/nginxoid/internal/timeutil"
)

var logRegex = regexp.MustCompile(`^(\S+)\s+-\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+(\S+)"\s+(\d+)\s+(\d+)\s+"([^"]*)"\s+"([^"]*)"$`)

// LogRecord represents a single parsed NGINX log line.
type LogRecord struct {
	RemoteAddr    string
	RemoteUser    string
	TimeLocal     time.Time
	Method        string
	Resource      string
	Protocol      string
	Status        int
	BodyBytesSent int64
	HTTPReferer   string
	HTTPUserAgent string
}

// ParseLine parses an NGINX log line into a LogRecord.
func ParseLine(line string) (*LogRecord, error) {
	matches := logRegex.FindStringSubmatch(line)
	if len(matches) != 11 {
		return nil, ErrMalformedLine
	}

	parsedTime, err := time.Parse(timeutil.NginxTimeLayout, matches[3])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse time %s: %w", ErrMalformedLine, matches[3], err)
	}

	status, err := strconv.Atoi(matches[7])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse status %s: %w", ErrMalformedLine, matches[7], err)
	}

	bodyBytes, err := strconv.ParseInt(matches[8], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse body bytes %s: %w", ErrMalformedLine, matches[8], err)
	}

	return &LogRecord{
		RemoteAddr:    matches[1],
		RemoteUser:    matches[2],
		TimeLocal:     parsedTime,
		Method:        matches[4],
		Resource:      matches[5],
		Protocol:      matches[6],
		Status:        status,
		BodyBytesSent: bodyBytes,
		HTTPReferer:   matches[9],
		HTTPUserAgent: matches[10],
	}, nil
}
