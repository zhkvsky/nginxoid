package report

import "errors"

// ErrReportAlreadyExists indicates that the report file already exists.
var ErrReportAlreadyExists = errors.New("report file already exists")
