package external

import (
	"io"
)

const ProgressByteLimit = 256 * 1024

// ProgressReader writes PROGRESS messages to an External as data is Read from
// it, at most once every ProgressByteLimit bytes.
type ProgressReader struct {
	r io.Reader
	e *External

	n         int64
	lastPrint int64
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
	if pr.n-pr.lastPrint > ProgressByteLimit || (pr.n != pr.lastPrint && err != nil) {
		pr.e.Progress(pr.n)
		pr.lastPrint = pr.n
	}
	return n, err
}
