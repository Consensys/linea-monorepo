//go:build !cuda

package gpu

// Enabled is false when the binary is built without the cuda tag.
// Use as a compile-time constant so the compiler eliminates dead branches:
//
//	if gpu.Enabled { /* GPU path */ } else { /* CPU path */ }
const Enabled = false
