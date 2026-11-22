package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zhkvsky/nginxoid/internal/analyzer"
	"github.com/zhkvsky/nginxoid/internal/config"
)

func renderJSON(rep Report) ([]byte, error) {
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal report to json: %w", err)
	}

	return data, nil
}

func renderMarkdown(rep Report) []byte {
	var b strings.Builder

	b.WriteString("# Отчет анализа логов\n\n")

	_, _ = fmt.Fprintf(&b, "- Количество файлов: %d\n", rep.FilesCount)
	_, _ = fmt.Fprintf(&b, "- Обработано строк: %d\n", rep.ProcessedLinesCount)
	_, _ = fmt.Fprintf(&b, "- Повреждённых строк: %d\n", rep.MalformedLinesCount)
	_, _ = fmt.Fprintf(&b, "- Всего запросов: %d\n\n", rep.TotalRequestsCount)

	b.WriteString("## Обработанные файлы\n\n")

	for _, file := range rep.Files {
		_, _ = fmt.Fprintf(&b, "- %s\n", file)
	}

	b.WriteString("\n## Размер ответа (в байтах)\n\n")
	_, _ = fmt.Fprintf(&b, "- Средний: %.2f\n", rep.ResponseSizeInBytes.Average)
	_, _ = fmt.Fprintf(&b, "- Максимальный: %d\n", rep.ResponseSizeInBytes.Max)
	_, _ = fmt.Fprintf(&b, "- 95-й перцентиль (p95): %d\n\n", rep.ResponseSizeInBytes.P95)

	b.WriteString("## Запрашиваемые ресурсы\n\n")
	b.WriteString("| Ресурс | Количество запросов |\n")
	b.WriteString("| --- | --- |\n")

	for _, item := range rep.Resources {
		_, _ = fmt.Fprintf(&b, "| %s | %d |\n", item.Resource, item.TotalRequestsCount)
	}

	b.WriteString("\n## Коды ответа\n\n")
	b.WriteString("| Код ответа | Количество |\n")
	b.WriteString("| --- | --- |\n")

	for _, code := range slices.Sorted(maps.Keys(rep.ResponseCodes)) {
		_, _ = fmt.Fprintf(&b, "| %d | %d |\n", code, rep.ResponseCodes[code])
	}

	return []byte(b.String())
}

func writeExclusive(outputPath string, data []byte) error {
	file, err := os.OpenFile(filepath.Clean(outputPath), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %s", ErrReportAlreadyExists, outputPath)
		}

		return fmt.Errorf("create report file: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("write report data: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close report file: %w", err)
	}

	return nil
}

// Write generates a report in the specified format and saves it to a file without overwriting existing files.
func Write(outputPath string, format config.Format, files []string, res analyzer.Result) error {
	rep := New(files, res)

	var (
		data []byte
		err  error
	)

	switch format {
	case config.FormatJSON:
		data, err = renderJSON(rep)
		if err != nil {
			return err
		}
	case config.FormatMarkdown:
		data = renderMarkdown(rep)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}

	return writeExclusive(outputPath, data)
}
