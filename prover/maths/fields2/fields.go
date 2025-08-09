package field

import "unsafe"

//go:generate go run ./internal/generator --config ./generator-config.json

// unsafeCast generically casts between pointer types. This is must be only used
// between types with the same layout.
func unsafeCast[U, V any](x *U) *V {
	return (*V)(unsafe.Pointer(x))
}
