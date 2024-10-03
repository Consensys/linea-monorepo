// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtle

import (
	"unsafe"
)

// XORBytes sets dst[i] = x[i] ^ y[i] for all i < n = min(len(x), len(y)),
// returning n, the number of bytes written to dst.
// If dst does not have length at least n,
// XORBytes panics without writing anything to dst.
func XORBytes(dst, x, y []byte) int {
	n := min(len(x), len(y))
	if n == 0 {
		return 0
	}
	if n > len(dst) {
		panic("subtle.XORBytes: dst too short")
	}
	xorBytes(&dst[0], &x[0], &y[0], n) // arch-specific
	return n
}

func XorIn(state *[25]uint64, buf [17]uint64) {
	x := (*[25 * 64 / 8]byte)(unsafe.Pointer(state))
	y := (*[17 * 64 / 8]byte)(unsafe.Pointer(&buf))
	XORBytes(x[:], x[:], y[:])
}

//
// xorIn xors the bytes in buf into the state.
//func xorIn(d *state, buf []byte) {
//	if cpu.IsBigEndian {
//		for i := 0; len(buf) >= 8; i++ {
//			a := binary.LittleEndian.Uint64(buf)
//			d.a[i] ^= a
//			buf = buf[8:]
//		}
//	} else {
//		ab := (*[25 * 64 / 8]byte)(unsafe.Pointer(&d.a))
//		XORBytes(ab[:], ab[:], buf)
//	}
//}
