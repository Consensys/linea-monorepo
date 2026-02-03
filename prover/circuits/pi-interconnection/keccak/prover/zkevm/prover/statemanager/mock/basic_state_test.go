package mock

import (
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/keccak"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/stretchr/testify/assert"
)

func TestBasicState(t *testing.T) {

	var (
		addA, _        = types.AddressFromHex("0xdeadbeafdeadbeafdeadbeafdeadbeafdeadbeaf")
		addB, _        = types.AddressFromHex("0xfede1234fede1234fede1234fede1234fede1234")
		balance        = big.NewInt(42)
		nonce    int64 = 47
		codeSize int64 = 67868
		codeHash       = types.DummyFullByte(32123)
		stoKeyA        = types.DummyFullByte(997)
		stoKeyB        = types.DummyFullByte(4569)
		stoValA        = types.DummyFullByte(21)
	)

	t.Run("GetSetBalance", func(t *testing.T) {
		state := State{}
		assert.Equal(t, state.GetBalance(addA), &big.Int{})
		state.SetBalance(addA, balance)
		assert.Equal(t, state.GetBalance(addA), balance)
		assert.Equal(t, state.GetBalance(addB), &big.Int{})
	})

	t.Run("GetSetNonce", func(t *testing.T) {
		state := State{}
		assert.Equal(t, state.GetNonce(addA), int64(0))
		state.SetNonce(addA, nonce)
		assert.Equal(t, state.GetNonce(addA), nonce)
		assert.Equal(t, state.GetNonce(addB), int64(0))
	})

	t.Run("GetSetCodeSize", func(t *testing.T) {
		state := State{}
		assert.Equal(t, state.GetCodeSize(addA), int64(0))
		state.SetCodeSize(addA, codeSize)
		assert.Equal(t, state.GetCodeSize(addA), codeSize)
		assert.Equal(t, state.GetCodeSize(addB), int64(0))
	})

	t.Run("GetSetCodeHash", func(t *testing.T) {
		state := State{}
		emptyHash := keccak.Hash([]byte{})
		assert.Equal(t, state.GetCodeHash(addA), types.AsFullBytes32(emptyHash[:]))
		state.SetCodeHash(addA, codeHash)
		assert.Equal(t, state.GetCodeHash(addA), codeHash)
		assert.Equal(t, state.GetCodeHash(addB), types.AsFullBytes32(emptyHash[:]))
	})

	t.Run("GetSetStorage", func(t *testing.T) {
		state := State{}
		assert.Equal(t, state.GetStorage(addA, stoKeyA), types.FullBytes32{})
		state.SetStorage(addA, stoKeyA, stoValA)
		assert.Equal(t, state.GetStorage(addA, stoKeyA), stoValA)
		assert.Equal(t, state.GetStorage(addB, stoKeyA), types.FullBytes32{})
		assert.Equal(t, state.GetStorage(addA, stoKeyB), types.FullBytes32{})
	})

	t.Run("GetSetStorage with deep-copy", func(t *testing.T) {
		state := State{}
		state.SetStorage(addA, stoKeyA, stoValA)

		stateCp := state.DeepCopy()
		assert.Equal(t, stateCp.GetStorage(addA, stoKeyA), stoValA)
		assert.Equal(t, stateCp.GetStorage(addB, stoKeyA), types.FullBytes32{})
		assert.Equal(t, stateCp.GetStorage(addA, stoKeyB), types.FullBytes32{})
	})

}
