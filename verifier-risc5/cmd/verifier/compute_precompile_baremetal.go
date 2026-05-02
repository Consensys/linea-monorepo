//go:build baremetal && zkvm_precompile

package main

import (
	"unsafe"
	_ "unsafe"

	"device/riscv"
	"runtime/volatile"

	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
)

//go:linkname zkvmPrecompileCall tinygo_zkvmPrecompile
func zkvmPrecompileCall(req uintptr) uintptr

func computeWordsImpl(words []uint64) (uint64, bool) {
	if len(words) > guestabi.PrecompileMaxWords {
		return 0, false
	}

	base := guestabi.PrecompileBase
	store32(base+0, guestabi.PrecompileMagic)
	store32(base+4, guestabi.PrecompileVersion)
	store32(base+8, guestabi.PrecompileOpcodeCompute)
	store32(base+12, guestabi.PrecompileStatusReady)
	store32(base+16, uint32(len(words)))
	store32(base+20, 0)
	store64(base+guestabi.PrecompileResultOffset, 0)

	for i, word := range words {
		store64(base+guestabi.PrecompileWordsOffset+uintptr(i)*8, word)
	}

	// Keep all precompile request stores visible before the host handles ECALL.
	riscv.Asm("fence rw, rw")
	status := uint32(zkvmPrecompileCall(base))
	riscv.Asm("fence rw, rw")
	if status != guestabi.PrecompileStatusSuccess {
		return 0, false
	}

	return load64(base + guestabi.PrecompileResultOffset), true
}

func store32(addr uintptr, value uint32) {
	volatile.StoreUint32((*uint32)(unsafe.Pointer(addr)), value)
}

func store64(addr uintptr, value uint64) {
	volatile.StoreUint64((*uint64)(unsafe.Pointer(addr)), value)
}

func load64(addr uintptr) uint64 {
	return volatile.LoadUint64((*uint64)(unsafe.Pointer(addr)))
}
