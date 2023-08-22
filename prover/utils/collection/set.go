package collection

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

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

// InsertNew inserts a new value and panics if it was contained
// already
func (kv *Set[K]) InsertNew(keys ...K) {
	// Check all keys are indeed news
	for _, key := range keys {
		if _, found := kv.inner[key]; found {
			utils.Panic("entry %v already inserted", key)
		}
	}
	// Then do the insertion
	for _, key := range keys {
		kv.inner[key] = struct{}{}
	}

}

// Returns the list of all the keys
func (kv Set[K]) ListAllKeys() []K {
	var res []K
	for k := range kv.inner {
		res = append(res, k)
	}
	return res
}

// Panic if the given entry does not exists
func (kv Set[K]) MustExists(keys ...K) {
	for _, key := range keys {
		if _, found := kv.inner[key]; !found {
			utils.Panic("entry %v does not exists", key)
		}
	}
}

// Iterates a function over all elements of the map
func (kv *Set[K]) IterateFunc(f func(k K)) {
	for k := range kv.inner {
		f(k)
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

// Returns the inner
func (kv Set[K]) Inner() map[K]struct{} {
	return kv.inner
}

// Delete an entry. Panic if the entry was not found
func (kv *Set[K]) Del(k K) {
	// Sanity-check
	kv.MustExists(k)
	delete(kv.inner, k)
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

/*
Pop a value from the set. Non-deterministic.
*/
func (kv *Set[K]) Pop() (k_ K, err error) {
	for k := range kv.inner {
		kv.Del(k)
		return k, nil
	}
	return k_, fmt.Errorf("emptied set")
}
