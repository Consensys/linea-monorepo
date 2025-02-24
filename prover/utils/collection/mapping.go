package collection

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// Mapping wraps a map and adds utility functions
type Mapping[K comparable, V any] struct {
	innerMap map[K]V
}

// Constructor for KVStore
func NewMapping[K comparable, V any]() Mapping[K, V] {
	return Mapping[K, V]{
		innerMap: make(map[K]V),
	}
}

// Attempts to retrieve a value from a given key. Panics
// if it fails
func (kv *Mapping[K, V]) MustGet(key K) V {
	res, found := kv.innerMap[key]

	if !found {
		utils.Panic("Entry %v does not exists", key)
	}

	return res
}

// Attempts to retrieve a value from a given key. Returns nil
// if it failed
func (kv *Mapping[K, V]) TryGet(key K) (V, bool) {
	res, ok := kv.innerMap[key]
	return res, ok
}

// InsertNew inserts a new value and panics if it was
// contained already
func (kv *Mapping[K, V]) InsertNew(key K, value V) {
	if _, found := kv.innerMap[key]; found {
		utils.Panic("Entry %v already found", key)
	}
	kv.innerMap[key] = value
}

// Update a key possibly overwriting any existing entry
func (kv *Mapping[K, V]) Update(key K, value V) {
	kv.innerMap[key] = value
}

// Returns the list of all the keys
func (kv *Mapping[K, V]) ListAllKeys() []K {
	res := make([]K, 0, len(kv.innerMap))
	for k := range kv.innerMap {
		res = append(res, k)
	}
	return res
}

// Panic if the given entry does not exists
func (kv *Mapping[K, V]) MustExists(keys ...K) {
	var missingListString error
	ok := true
	for _, key := range keys {
		if _, found := kv.innerMap[key]; !found {

			// accumulate the keys in an user-friendly error message
			if missingListString == nil {
				missingListString = fmt.Errorf("%v", key)
			} else {
				missingListString = fmt.Errorf("%v, %v", missingListString, key)
			}

			ok = false
		}
	}
	if !ok {
		utils.Panic("MustExists : assertion failed. (%v are missing)", missingListString)
	}
}

// Iterates a function over all elements of the map
func (kv *Mapping[K, V]) IterateFunc(f func(k K, v V)) {
	for k, v := range kv.innerMap {
		f(k, v)
	}
}

// Returns `true` if all the passed entries exists
func (kv *Mapping[K, V]) Exists(ks ...K) bool {
	for _, k := range ks {
		_, found := kv.innerMap[k]
		if !found {
			return false
		}
	}
	return true
}

// ToSlice lists all entries in a slice of tuple
func (kv *Mapping[K, V]) ListValues() []V {
	res := make([]V, 0, len(kv.innerMap))
	for _, v := range kv.innerMap {
		res = append(res, v)
	}
	return res
}

// Returns the innerMap
func (kv *Mapping[K, V]) InnerMap() map[K]V {
	return kv.innerMap
}

// Delete an entry. Panic if the entry was not found
func (kv *Mapping[K, V]) Del(k K) {
	// Sanity-check
	kv.MustGet(k)
	delete(kv.innerMap, k)
}

// Delete an entry. NOOP if the entry was not found
func (kv *Mapping[K, V]) TryDel(k K) bool {
	// Sanity-check
	found := kv.Exists(k)
	if found {
		kv.MustGet(k)
		delete(kv.innerMap, k)
	}
	return found
}
