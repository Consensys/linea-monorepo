package utils

import "sync/atomic"

// Scratch is a simple scratch space allocator for vectors of fixed size.
// It is safe for concurrent use by multiple goroutines.
type Scratch[T any] struct {
	s          []T
	offset     int64
	vectorSize int
}

// NewScratch creates a new Scratch space that can hold nbVectors of size vectorSize.
func NewScratch[T any](vectorSize, nbVectors int) *Scratch[T] {
	return &Scratch[T]{
		s:          make([]T, vectorSize*nbVectors),
		offset:     0,
		vectorSize: vectorSize,
	}
}

// NewVector returns a new vector of size vectorSize from the scratch space.
// If the scratch space is exhausted, it allocates a new vector.
// It is safe for concurrent use by multiple goroutines.
func (s *Scratch[T]) NewVector() []T {
	n := atomic.AddInt64(&s.offset, 1)
	start := (n - 1) * int64(s.vectorSize)
	end := n * int64(s.vectorSize)
	if end > int64(len(s.s)) {
		return make([]T, s.vectorSize)
	}
	return s.s[start:end]
}

// Reset resets the scratch space offset to zero.
func (s *Scratch[T]) Reset() {
	s.offset = 0
}
