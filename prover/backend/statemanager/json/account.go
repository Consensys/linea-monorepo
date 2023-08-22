package json

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/valyala/fastjson"
)

// Sugar type aliasing
type (
	Account = eth.Account
)

// try parsing an hex encoded account in a json
func tryParseAccount(p fastjson.Value, field ...string) (Account, error) {

	_p := p.Get(field...)
	if _p == nil {
		return Account{}, fmt.Errorf("not found : %v", field)
	}

	// Safe to dereference
	p = *_p

	// Attempt to parse the hexString
	b, err := jsonutil.TryGetHexBytes(p)
	if err != nil {
		return Account{}, err
	}

	/*
		Now the account should have the following shape

		nonce, balance, storageRoot, mimcCodeHash, keccakCodeHash, codeSize

		0000000000000000000000000000000000000000000000000000000000000041 // nonce
		0000000000000000000000000000000000000000000000000000000000000343 // balance
		28aed60bedfcad80c2a5e6a7a3100e837f875f9aa71d768291f68f894b0a3d11 // storageRoot
		2c7298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6ee // mimcCodeHash
		c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470 // keccakCodeHash
		0000000000000000000000000000000000000000000000000000000000000000 // codesize
	*/
	acc := Account{}
	buf := bytes.Buffer{}
	buf.Write(b)
	word := [32]byte{}

	buf.Read(word[:])
	acc.Nonce, err = asInt64(word[:])
	if err != nil {
		return Account{}, err
	}

	buf.Read(word[:])
	acc.Balance = &big.Int{} // initialize the bigint before setting
	acc.Balance.SetBytes(word[:])

	buf.Read(word[:])
	acc.StorageRoot = Digest(word)

	buf.Read(word[:])
	acc.CodeHash = Digest(word)

	buf.Read(word[:])
	acc.KeccakCodeHash = jsonutil.Bytes32(word)

	buf.Read(word[:])
	acc.CodeSize, err = asInt64(word[:])
	if err != nil {
		return Account{}, err
	}

	return acc, nil
}

func asInt64(b []byte) (int64, error) {
	if len(b) != 32 {
		panic("BUG: should be 32 bytes")
	}

	var l big.Int
	l.SetBytes(b)

	if !l.IsInt64() {
		return 0, fmt.Errorf("not an int64")
	}

	return l.Int64(), nil
}
