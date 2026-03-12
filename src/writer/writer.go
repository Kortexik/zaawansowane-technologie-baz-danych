package writer

import (
	"os"
	"sync"
)

type BatchWriter struct {
	file  *os.File
	buf   []byte
	size  int
	count int
	mu    sync.Mutex
}

func New(file *os.File, size int) *BatchWriter {
	return &BatchWriter{
		file: file,
		buf:  make([]byte, 0, size*256),
		size: size,
	}
}

func (bw *BatchWriter) Write(data string) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.buf = append(bw.buf, data...)
	bw.count++

	if bw.count >= bw.size {
		return bw.flush()
	}
	return nil
}

func (bw *BatchWriter) flush() error {
	if len(bw.buf) == 0 {
		return nil
	}
	_, err := bw.file.Write(bw.buf)
	bw.buf = bw.buf[:0]
	bw.count = 0
	return err
}

func (bw *BatchWriter) Close() error {
	_ = bw.flush()
	return bw.file.Close()
}
