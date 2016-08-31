package external

import (
	"io"
	"time"
)

const ProgressTimeInterval = time.Second / 2

// ProgressReader writes PROGRESS messages to an External as data is Read from
// it, at most once every ProgressTimeInterval.
type ProgressReader struct {
	r io.Reader
	e *External

	n          int64
	lastPrintN int64
	lastPrint  time.Time
}

func NewProgressReader(r io.Reader, e *External) *ProgressReader {
	return &ProgressReader{
		r: r,
		e: e,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.n += int64(n)
	if time.Since(pr.lastPrint) > ProgressTimeInterval ||
		(err != nil && pr.n != pr.lastPrintN) {

		pr.e.Progress(pr.n)
		pr.lastPrintN = pr.n
		pr.lastPrint = time.Now()
	}
	return n, err
}
