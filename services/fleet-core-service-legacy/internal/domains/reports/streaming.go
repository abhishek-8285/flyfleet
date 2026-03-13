package reports

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	MaxReportRows = 5000000
	MaxReportSize = 2147483648 // 2GB
)

var (
	ErrRowLimitExceeded  = fmt.Errorf("report row limit exceeded")
	ErrFileSizeExceeded = fmt.Errorf("report file size exceeded")
)

// byteCounter tracks the number of bytes written to it.
type byteCounter struct {
	count int64
}

func (bc *byteCounter) Write(p []byte) (n int, err error) {
	bc.count += int64(len(p))
	return len(p), nil
}

// StreamingCSVWriter writes CSV rows in chunks to avoid full-dataset memory accumulation.
// Spec: cursor-based query, row streaming, file streaming (Section 14).
type StreamingCSVWriter struct {
	storage      ObjectStorage
	chunkSize    int // rows per chunk (default 10000 per spec)
	compressed   bool
	logger       *zap.Logger
	totalBytes   int64
}

func NewStreamingCSVWriter(storage ObjectStorage, logger *zap.Logger) *StreamingCSVWriter {
	return &StreamingCSVWriter{
		storage:    storage,
		chunkSize:  10000,
		compressed: true,
		logger:     logger,
	}
}

// WriteAndUpload streams rows from the channel, writes them as gzipped CSV in chunks,
// and uploads to object storage. Returns the file URL and total row count.
func (w *StreamingCSVWriter) WriteAndUpload(
	ctx context.Context,
	reportID string,
	headers []string,
	rowCh <-chan []string,
) (fileURL string, rowCount int64, err error) {

	w.totalBytes = 0
	var buf bytes.Buffer
	var gw *gzip.Writer
	var csvW *csv.Writer

	reset := func() {
		buf.Reset()
		// To track bytes precisely, we'll use a custom writer that counts.
		// But gzip.Writer writes to buf. We want to know how much is in buf after Close.
		// Actually, we can just check buf.Len() before each upload.
		gw = gzip.NewWriter(&buf)
		csvW = csv.NewWriter(gw)
	}
	reset()

	writeHeaders := func() error {
		return csvW.Write(headers)
	}

	flushAndUpload := func(part int) (string, error) {
		csvW.Flush()
		if err := csvW.Error(); err != nil {
			return "", fmt.Errorf("csv flush error: %w", err)
		}
		if err := gw.Close(); err != nil {
			return "", fmt.Errorf("gzip close error: %w", err)
		}

		w.totalBytes += int64(buf.Len())
		if w.totalBytes > MaxReportSize {
			return "", ErrFileSizeExceeded
		}

		key := fmt.Sprintf("%s/part-%03d.csv.gz", reportID, part)
		url, err := w.storage.Upload(ctx, key, &buf, "application/gzip")
		if err != nil {
			return "", fmt.Errorf("upload error: %w", err)
		}
		return url, nil
	}

	if err := writeHeaders(); err != nil {
		return "", 0, err
	}

	part := 0
	chunkRows := 0
	var lastURL string

	for row := range rowCh {
		select {
		case <-ctx.Done():
			return "", rowCount, ctx.Err()
		default:
		}

		if err := csvW.Write(row); err != nil {
			return "", rowCount, fmt.Errorf("csv write error: %w", err)
		}
		rowCount++
		chunkRows++

		if rowCount > MaxReportRows {
			return "", rowCount, ErrRowLimitExceeded
		}

		if chunkRows >= w.chunkSize {
			url, err := flushAndUpload(part)
			if err != nil {
				return "", rowCount, err
			}
			lastURL = url
			part++
			chunkRows = 0
			reset()
			writeHeaders()
			w.logger.Info("CSV part uploaded",
				zap.String("report_id", reportID),
				zap.Int("part", part),
				zap.Int64("total_rows", rowCount),
				zap.Int64("total_bytes", w.totalBytes))
		}
	}

	// Upload final chunk
	if chunkRows > 0 {
		url, err := flushAndUpload(part)
		if err != nil {
			return "", rowCount, err
		}
		lastURL = url
	}

	w.logger.Info("Streaming CSV upload complete",
		zap.String("report_id", reportID),
		zap.Int("parts", part+1),
		zap.Int64("total_rows", rowCount),
		zap.Int64("total_bytes", w.totalBytes))

	// Return the last part URL (or primary manifest URL in prod)
	return lastURL, rowCount, nil
}

// MockRowSource generates fake rows for a given report type (dev mode).
// In production, replace with real aggregation queries against daily_*_metrics tables.
func MockRowSource(ctx context.Context, reportType ReportType, params map[string]interface{}, onProgress func(int)) <-chan []string {
	ch := make(chan []string, 100)
	rowCount := 50 // Simulate 50 rows for dev

	go func() {
		defer close(ch)
		for i := 0; i < rowCount; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}
			ch <- []string{
				fmt.Sprintf("row-%d", i),
				string(reportType),
				time.Now().Format(time.RFC3339),
				"tenant-data",
			}
			if onProgress != nil && i%10 == 0 {
				onProgress(int(float64(i) / float64(rowCount) * 100))
			}
			time.Sleep(10 * time.Millisecond) // simulate query latency
		}
	}()
	return ch
}
