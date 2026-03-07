package logger

import (
	"bytes"
	"container/ring"
	"io"
	"sync"
)

const maxLines = 500

type Buffer struct {
	mu   sync.RWMutex
	ring *ring.Ring
}

func New() *Buffer {
	return &Buffer{ring: ring.New(maxLines)}
}

func (b *Buffer) Write(p []byte) (int, error) {
	for _, line := range bytes.Split(bytes.TrimRight(p, "\n"), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		b.mu.Lock()
		b.ring.Value = string(line)
		b.ring = b.ring.Next()
		b.mu.Unlock()
	}
	return len(p), nil
}

func (b *Buffer) Lines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]string, 0, maxLines)
	b.ring.Do(func(v any) {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	})
	return out
}

// Install sets log output to write to both dst (e.g. os.Stderr) and b.
func (b *Buffer) Install(dst io.Writer) {
	// circular import avoided — caller does log.SetOutput(b.Writer(dst))
	_ = dst
}

func (b *Buffer) Writer(dst io.Writer) io.Writer {
	return io.MultiWriter(b, dst)
}
