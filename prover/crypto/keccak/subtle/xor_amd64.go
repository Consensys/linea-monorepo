//go:build !purego

package subtle

//go:noescape
func xorBytes(dst, a, b *byte, n int)
