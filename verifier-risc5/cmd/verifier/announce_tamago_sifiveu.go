//go:build baremetal && tamago_sifive_u

package main

import "github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"

func announceBaremetalResult(value uint64) {
	writeGuestStatus(guestabi.StatusCodeSuccess, value, value)
}

func announceBaremetalMismatch(expected, got uint64) {
	writeGuestStatus(guestabi.StatusCodeMismatch, got, expected)
}

func announceBaremetalInputError() {
	writeGuestStatus(guestabi.StatusCodeInputError, 0, 0)
}
