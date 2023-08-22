package collection

// This structure accounts for the ordering of the commitment
// in the interactions with the verifier
type VecSet[T comparable] struct {
	inner []Set[T]
}

// Construct a new empty VecSet
func NewVecSet[T comparable]() VecSet[T] {
	return VecSet[T]{inner: make([]Set[T], 0)}
}

// Insert registers a new commitment name into the sequence at
// a given round number.
func (s *VecSet[T]) InsertAt(pos int, t ...T) {
	s.Reserve(pos + 1)
	s.inner[pos].InsertNew(t...)
}

// Returns all the names at a given round
func (s *VecSet[T]) Get(pos int) Set[T] {
	return s.inner[pos]
}

// Allocates up to a given rounds number
func (s *VecSet[T]) Reserve(newLen int) {
	// We may not have to append the sequence
	// If we need to, we append to it as many time as we need
	for len(s.inner) < newLen {
		s.inner = append(s.inner, NewSet[T]())
	}
}

/*
Returns the length of a VecSet
*/
func (s *VecSet[T]) Len() int {
	return len(s.inner)
}

/*
Delete an entry, which is expected to be at a given round.
Panic if the entry is missing.
*/
func (s *VecSet[T]) Del(pos int, t T) {
	s.inner[pos].Del(t)
}
