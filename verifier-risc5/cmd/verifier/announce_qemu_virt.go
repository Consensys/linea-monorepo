//go:build baremetal && qemu_virt

package main

import "unsafe"

const (
	qemuVirtUARTBase = uintptr(0x10000000)
	uartTHROffset    = uintptr(0)
	uartLSROffset    = uintptr(5)
	uartLSRTHRE      = 0x20
)

func announceBaremetal(value uint64) {
	uartWriteString("verifier result 0x")
	uartWriteHex64(value)
	uartWriteString("\r\n")
}

func uartWriteString(s string) {
	for i := 0; i < len(s); i++ {
		uartWriteByte(s[i])
	}
}

func uartWriteHex64(value uint64) {
	const hex = "0123456789abcdef"

	for shift := uint(60); shift < 64; shift -= 4 {
		nibble := byte((value >> shift) & 0x0f)
		uartWriteByte(hex[nibble])
		if shift == 0 {
			return
		}
	}
}

func uartWriteByte(value byte) {
	for (*(*uint8)(unsafe.Pointer(qemuVirtUARTBase + uartLSROffset)) & uartLSRTHRE) == 0 {
	}

	*(*uint8)(unsafe.Pointer(qemuVirtUARTBase + uartTHROffset)) = value
}
