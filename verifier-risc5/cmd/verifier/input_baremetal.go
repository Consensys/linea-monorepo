//go:build baremetal

package main

import (
	"unsafe"

	"github.com/consensys/linea-monorepo/verifier-risc5/internal/guestabi"
)

func loadVerifierInput() (verifierInput, bool) {
	header := (*guestabi.Header)(unsafe.Pointer(guestabi.InputBase))
	if header.Magic != guestabi.Magic || header.Version != guestabi.Version {
		return verifierInput{}, false
	}

	count := int(header.WordCount)
	if count > guestabi.MaxWords {
		return verifierInput{}, false
	}

	words := unsafe.Slice((*uint64)(unsafe.Pointer(guestabi.InputBase+guestabi.HeaderSize)), count)
	return verifierInput{
		Words:    words,
		Expected: header.Expected,
	}, true
}
