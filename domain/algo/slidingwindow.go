package algo

type SlidingWindow[T any] struct {
	maxLen int
	buffer []T
}

func NewSlidingWindow[T any](maxLen int) SlidingWindow[T] {
	return SlidingWindow[T]{maxLen: maxLen}
}

func (w *SlidingWindow[T]) MaxLen() int {
	return w.maxLen
}

func (w *SlidingWindow[T]) Add(value T) {
	if len(w.buffer) == 2*w.maxLen {
		copy(w.buffer[:w.maxLen], w.buffer[w.maxLen:])
		w.buffer = w.buffer[:w.maxLen]
	}
	w.buffer = append(w.buffer, value)
}

func (w *SlidingWindow[T]) Len() int {
	return min(w.maxLen, len(w.buffer))
}

func (w *SlidingWindow[T]) Item(index int) T {
	var start = max(0, len(w.buffer)-w.maxLen)
	return w.buffer[start+index]
}

func (w *SlidingWindow[T]) Items() []T {
	var start = max(0, len(w.buffer)-w.maxLen)
	return w.buffer[start:]
}
