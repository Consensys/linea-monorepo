//go:build baremetal

package main

import (
	"unsafe"

	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
)

var guestStatus = (*guestabi.Status)(unsafe.Pointer(guestabi.StatusBase))

func writeGuestStatus(code uint32, result, expected uint64) {
	guestStatus.Magic = guestabi.StatusMagic
	guestStatus.Version = guestabi.StatusVersion
	guestStatus.Code = code
	guestStatus.Reserved = 0
	guestStatus.Result = result
	guestStatus.Expected = expected
}
