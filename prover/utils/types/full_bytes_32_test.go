package types_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/consensys/linea-monorepo/prover/utils/types"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
)

func TestFullBytes32(t *testing.T) {
	val := FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	written := "0x0000c5d200004601000086f70000233c0000927e00007db20000dcc7000003c00000e5000000b6530000ca820000273b00007bfa0000d80400005d850000a470"
	fullBytes32Hash := "0x76f5f775402233c90584827a772ac4fa1be76ee2496d5964431708c573e2bd32"

	buf := &bytes.Buffer{}
	val.WriteTo(buf)
	written2 := utils.HexEncodeToString(buf.Bytes())
	assert.Equal(t, written, written2)

	// Calculate the hash
	hasher := poseidon2.Poseidon2()
	val.WriteTo(hasher)
	hash := hasher.Sum(nil)

	// Format the hash as a "0x" prefixed hex string
	hashString := fmt.Sprintf("0x%x", hash)

	assert.Equal(t, fullBytes32Hash, hashString)

}
