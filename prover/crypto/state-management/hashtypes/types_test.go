package hashtypes_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/stretchr/testify/assert"
)

func TestWriteBigInt(t *testing.T) {

	b := big.NewInt(10)

	buffer := &bytes.Buffer{}
	hashtypes.WriteBigIntTo(buffer, b)

	// Converts to hex to simplify the reading
	hex := hex.EncodeToString(buffer.Bytes())
	assert.Equal(
		t,
		"000000000000000000000000000000000000000000000000000000000000000a",
		hex,
	)
}

func TestWriteInt64(t *testing.T) {

	n := int64(10)

	buffer := &bytes.Buffer{}
	hashtypes.WriteInt64To(buffer, n)

	// Converts to hex to simplify the reading
	hex := hex.EncodeToString(buffer.Bytes())
	assert.Equal(
		t,
		"000000000000000000000000000000000000000000000000000000000000000a",
		hex,
	)
}
