package notify

// BufferLen exposes the current buffer length for white-box testing.
func BufferLen(b *BufferedNotifier) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}
