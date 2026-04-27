//go:build cuda

package gpu

// Enabled is true when the binary is built with the cuda tag.
// Use as a compile-time constant so the compiler eliminates dead branches:
//
//	if gpu.Enabled { /* GPU path */ } else { /* CPU path */ }
const Enabled = true
