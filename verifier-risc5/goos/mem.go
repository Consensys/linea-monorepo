//go:build tamago && riscv64 && tamago_libriscv && !tiny

package goos

const (
	ArenaBaseOffset     = 0
	HeapAddrBits        = 40
	LogHeapArenaBytes   = 2 + 20
	LogPallocChunkPages = 9
	MinPhysPageSize     = 4096
	StackSystem         = 0
)
