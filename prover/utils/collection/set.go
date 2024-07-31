package collection

// A set is an unordered collection addressed by keys, which supports
type Set[T comparable] struct {
	inner map[T]struct{}
}

// Constructor for KVStore
func NewSet[K comparable]() Set[K] {
	return Set[K]{
		inner: make(map[K]struct{}),
	}
}

// Returns `true` if the entry exists
func (kv Set[K]) Exists(ks ...K) bool {
	for _, k := range ks {
		_, found := kv.inner[k]
		if !found {
			return false
		}
	}
	return true
}

/*
Inserts regarless of whether the entry was already
present of not. Returns whether the entry was present
already.
*/
func (kv *Set[K]) Insert(k K) bool {
	if _, ok := kv.inner[k]; !ok {
		kv.inner[k] = struct{}{}
		return false
	}
	return true
}
