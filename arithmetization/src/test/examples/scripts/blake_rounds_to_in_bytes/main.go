// Generates a synthetic Blake-2f `IN_BYTES` test vector with a chosen number of
// compression rounds — useful for benchmarking the zkc interpreter at a
// configurable workload size on a single Blake vector.
//
// The h, m, t, f fields are hard-coded to the canonical Blake2b "abc" test
// vector (matches every line of go-corset's testdata/zkc/bench/blake.accepts —
// they all share the same h, m, t and the most common f=1 = final block).
// Only the rounds count varies. The expected h_out is computed live by running
// the Blake2b-F compression function in Go, so the generated line is valid
// against rust/src/blake/blake_with_in_bytes.rs for any rounds value the user
// passes.
//
// Output: one `0x<hex>` line on stdout (277 bytes = 554 hex chars), matching
// the .all format consumed by the blake_accepts_to_in_bytes helper / the
// blake-rust-json Makefile target.
//
// Usage:
//
//	blake_rounds_to_in_bytes <rounds>
//
// Reference: RFC 7693 / EIP-152. Cross-checked against the rounds=12 test
// vector documented in blake_with_in_bytes.rs (canonical Blake2b-512("abc")).
package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

// Standard Blake2b-512 initial state for the "abc" test:
//
//	h[0] = IV[0] XOR 0x01010040 (param block: 64-byte digest, 0 keylen, fanout=depth=1).
//	h[1..7] = IV[1..7] verbatim.
var initH = [8]uint64{
	0x6a09e667f2bdc948,
	0xbb67ae8584caa73b,
	0x3c6ef372fe94f82b,
	0xa54ff53a5f1d36f1,
	0x510e527fade682d1,
	0x9b05688c2b3e6c1f,
	0x1f83d9abfb41bd6b,
	0x5be0cd19137e2179,
}

// "abc" packed into m[0] as little-endian bytes (61 62 63 then 125 zero bytes).
var initM = [16]uint64{0x636261, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// t = (3, 0): message length = 3 bytes.
var initT = [2]uint64{3, 0}

// Final-block flag set (matches the most common case in blake.accepts).
const initF byte = 1

// Blake2b IV — unmixed initial state.
var iv = [8]uint64{
	0x6a09e667f3bcc908,
	0xbb67ae8584caa73b,
	0x3c6ef372fe94f82b,
	0xa54ff53a5f1d36f1,
	0x510e527fade682d1,
	0x9b05688c2b3e6c1f,
	0x1f83d9abfb41bd6b,
	0x5be0cd19137e2179,
}

// SIGMA permutations (matches blake_core.rs).
var sigma = [10][16]int{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
	{12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11},
	{13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10},
	{6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5},
	{10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0},
}

func rotr64(x uint64, n uint) uint64 {
	return (x >> n) | (x << (64 - n))
}

func mix(v *[16]uint64, a, b, c, d int, x, y uint64) {
	v[a] = v[a] + v[b] + x
	v[d] = rotr64(v[d]^v[a], 32)
	v[c] = v[c] + v[d]
	v[b] = rotr64(v[b]^v[c], 24)
	v[a] = v[a] + v[b] + y
	v[d] = rotr64(v[d]^v[a], 16)
	v[c] = v[c] + v[d]
	v[b] = rotr64(v[b]^v[c], 63)
}

// Blake2b-F compression with configurable rounds (EIP-152 form).
// Returns the new h state as 8 u64s.
func compress(rounds uint32, h [8]uint64, m [16]uint64, t [2]uint64, f byte) [8]uint64 {
	var v [16]uint64
	copy(v[:8], h[:])
	copy(v[8:], iv[:])
	v[12] ^= t[0]
	v[13] ^= t[1]
	if f != 0 {
		v[14] ^= ^uint64(0)
	}
	for i := uint32(0); i < rounds; i++ {
		s := sigma[i%10]
		mix(&v, 0, 4, 8, 12, m[s[0]], m[s[1]])
		mix(&v, 1, 5, 9, 13, m[s[2]], m[s[3]])
		mix(&v, 2, 6, 10, 14, m[s[4]], m[s[5]])
		mix(&v, 3, 7, 11, 15, m[s[6]], m[s[7]])
		mix(&v, 0, 5, 10, 15, m[s[8]], m[s[9]])
		mix(&v, 1, 6, 11, 12, m[s[10]], m[s[11]])
		mix(&v, 2, 7, 8, 13, m[s[12]], m[s[13]])
		mix(&v, 3, 4, 9, 14, m[s[14]], m[s[15]])
	}
	for i := 0; i < 8; i++ {
		h[i] ^= v[i] ^ v[i+8]
	}
	return h
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: blake_rounds_to_in_bytes <rounds>")
		os.Exit(1)
	}
	roundsU64, err := strconv.ParseUint(os.Args[1], 0, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: rounds must be a uint32: %v\n", err)
		os.Exit(1)
	}
	rounds := uint32(roundsU64)

	hOut := compress(rounds, initH, initM, initT, initF)

	var buf [277]byte
	binary.BigEndian.PutUint32(buf[0:4], rounds)
	for i := 0; i < 8; i++ {
		binary.LittleEndian.PutUint64(buf[4+i*8:], initH[i])
	}
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint64(buf[68+i*8:], initM[i])
	}
	binary.LittleEndian.PutUint64(buf[196:], initT[0])
	binary.LittleEndian.PutUint64(buf[204:], initT[1])
	buf[212] = initF
	for i := 0; i < 8; i++ {
		binary.LittleEndian.PutUint64(buf[213+i*8:], hOut[i])
	}

	fmt.Printf("0x%s\n", hex.EncodeToString(buf[:]))
	fmt.Fprintf(os.Stderr, "rounds=%d h_out (LE bytes hex): %s\n",
		rounds, hex.EncodeToString(buf[213:]))
}
