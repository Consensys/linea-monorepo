//go:build baremetal && qemu_virt

package main

import (
	"unsafe"

	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
)

func announceBaremetalResult(value uint64) {
	writeGuestStatus(guestabi.StatusCodeSuccess, value, value)
	qemuVirtExit(guestabi.StatusCodeSuccess)
}

func announceBaremetalMismatch(expected, got uint64) {
	writeGuestStatus(guestabi.StatusCodeMismatch, got, expected)
	qemuVirtExit(guestabi.StatusCodeMismatch)
}

func announceBaremetalInputError() {
	writeGuestStatus(guestabi.StatusCodeInputError, 0, 0)
	qemuVirtExit(guestabi.StatusCodeInputError)
}

func qemuVirtExit(code uint32) {
	value := guestabi.QEMUTestPass
	if code != guestabi.StatusCodeSuccess {
		value = (code << 16) | guestabi.QEMUTestFail
	}

	*(*uint32)(unsafe.Pointer(guestabi.QEMUTestBase)) = value
	haltForever()
}
