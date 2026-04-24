//go:build baremetal && tamago_libriscv

package main

import (
	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
)

func writeQEMUTest(value uint32)

func announceBaremetalResult(value uint64) {
	writeGuestStatus(guestabi.StatusCodeSuccess, value, value)
	tamagoLibriscvExit(guestabi.StatusCodeSuccess)
}

func announceBaremetalMismatch(expected, got uint64) {
	writeGuestStatus(guestabi.StatusCodeMismatch, got, expected)
	tamagoLibriscvExit(guestabi.StatusCodeMismatch)
}

func announceBaremetalInputError() {
	writeGuestStatus(guestabi.StatusCodeInputError, 0, 0)
	tamagoLibriscvExit(guestabi.StatusCodeInputError)
}

func tamagoLibriscvExit(code uint32) {
	value := guestabi.QEMUTestPass
	if code != guestabi.StatusCodeSuccess {
		value = (code << 16) | guestabi.QEMUTestFail
	}

	writeQEMUTest(value)
	haltForever()
}
