//go:build tamago && riscv64 && tamago_libriscv && tiny

package goos

const (
	ArenaBaseOffset     = 0
	HeapAddrBits        = 32
	LogHeapArenaBytes   = 9 + 10
	LogPallocChunkPages = 6
	MinPhysPageSize     = 96
	StackSystem         = 0
)
