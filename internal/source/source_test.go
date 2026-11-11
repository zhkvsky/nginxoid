package source

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	content := "line 1\nline 2\n"
	err := os.WriteFile(filePath, []byte(content), 0o600)
	require.NoError(t, err)

	src, err := Open(t.Context(), filePath)
	require.NoError(t, err)

	require.NotNil(t, src)
	defer func() {
		_ = src.Reader.Close()
	}()

	assert.Equal(t, []string{filePath}, src.Files)

	data, err := io.ReadAll(src.Reader)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestOpen_Glob(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "1.log")
	err := os.WriteFile(file1, []byte("file 1\n"), 0o600)
	require.NoError(t, err)

	file2 := filepath.Join(tmpDir, "2.log")
	err = os.WriteFile(file2, []byte("file 2\n"), 0o600)
	require.NoError(t, err)

	pattern := filepath.Join(tmpDir, "*.log")

	src, err := Open(t.Context(), pattern)
	require.NoError(t, err)

	require.NotNil(t, src)
	defer func() {
		_ = src.Reader.Close()
	}()

	assert.Equal(t, []string{file1, file2}, src.Files)

	data, err := io.ReadAll(src.Reader)
	require.NoError(t, err)
	assert.Equal(t, "file 1\nfile 2\n", string(data))
}

func TestOpen_URL(t *testing.T) {
	expectedBody := "log from http\n"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	}))
	defer ts.Close()

	src, err := Open(t.Context(), ts.URL)
	require.NoError(t, err)

	require.NotNil(t, src)
	defer func() {
		_ = src.Reader.Close()
	}()

	assert.Equal(t, []string{ts.URL}, src.Files)

	data, err := io.ReadAll(src.Reader)
	require.NoError(t, err)
	assert.Equal(t, expectedBody, string(data))
}

func TestOpen_Errors(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		missingPath := filepath.Join(tmpDir, "missing.log")

		src, err := Open(t.Context(), missingPath)
		require.Error(t, err)
		assert.Nil(t, src)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty glob pattern", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyPattern := filepath.Join(tmpDir, "*.nonexistent")

		src, err := Open(t.Context(), emptyPattern)
		require.Error(t, err)
		assert.Nil(t, src)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("http 404 not found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		src, err := Open(t.Context(), ts.URL)
		require.Error(t, err)
		assert.Nil(t, src)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
