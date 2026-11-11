// Package source handles reading log streams from local filesystem paths and remote URLs.
package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Source represents an opened data source for streaming log reads.
type Source struct {
	Files  []string
	Reader io.ReadCloser
}

// Open opens local files matching a path or glob pattern, or fetches a remote stream via URL.
func Open(ctx context.Context, path string) (*Source, error) {
	isURL := strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")

	if isURL {
		return openURL(ctx, path)
	}

	return openFiles(path)
}

func openURL(ctx context.Context, rawURL string) (*Source, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %w", ErrNotFound, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch url: %w", ErrNotFound, err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%w: remote server returned status %d", ErrNotFound, resp.StatusCode)
	}

	return &Source{
		Files:  []string{rawURL},
		Reader: resp.Body,
	}, nil
}

type multiReadCloser struct {
	reader  io.Reader
	closers []io.Closer
}

func (m *multiReadCloser) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func (m *multiReadCloser) Close() error {
	var firstErr error
	for _, c := range m.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func openFiles(pattern string) (*Source, error) {
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid glob pattern %q: %w", ErrNotFound, pattern, err)
	}

	if len(filenames) == 0 {
		return nil, fmt.Errorf("%w: no files found matching pattern %q", ErrNotFound, pattern)
	}

	readers := make([]io.Reader, 0, len(filenames))
	closers := make([]io.Closer, 0, len(filenames))

	for _, path := range filenames {
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			for _, c := range closers {
				_ = c.Close()
			}

			return nil, fmt.Errorf("%w: failed to open file %s: %w", ErrNotFound, path, err)
		}

		readers = append(readers, file)
		closers = append(closers, file)
	}

	return &Source{
		Files: filenames,
		Reader: &multiReadCloser{
			reader:  io.MultiReader(readers...),
			closers: closers,
		},
	}, nil
}
