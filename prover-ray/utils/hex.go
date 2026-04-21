// Package utils provides general-purpose utility functions used across the prover.
package utils

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	bls12377fr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// HexHashUint64 encodes the uint64 into an hexstring representing it as a u256 in bigendian form.
func HexHashUint64(v ...uint64) string {
	buffer := bytes.Buffer{}
	var I big.Int
	for i := range v {
		I.SetUint64(v[i])
		bytes := I.Bytes()
		bytes = append(make([]byte, 32-len(bytes)), bytes...)
		buffer.Write(bytes)
	}
	res := crypto.Keccak256(buffer.Bytes())
	return hexutil.Encode(res)
}

// FmtInt32Bytes formats an integer v as a 32-byte big-endian representation.
func FmtInt32Bytes(v int) [32]byte {
	var res [32]byte
	b := big.NewInt(int64(v)).Bytes()
	copy(res[32-len(b):], b)
	return res
}

// FmtIntHex32Bytes formats an integer as a 32 bytes hex string.
func FmtIntHex32Bytes(v int) string {
	bytes := FmtInt32Bytes(v)
	return hexutil.Encode(bytes[:])
}

// ApplyModulusBn254 applies the modulus of the BN254 scalar field.
func ApplyModulusBn254(b string) string {
	var f bn254fr.Element
	if _, err := f.SetString(b); err != nil {
		panic(err)
	}
	fbytes := f.Bytes()
	return hexutil.Encode(fbytes[:])
}

// ApplyModulusBls12377 applies the modulus of the BLS12-377 scalar field.
func ApplyModulusBls12377(b string) string {
	var f bls12377fr.Element
	if _, err := f.SetString(b); err != nil {
		panic(err)
	}
	fbytes := f.Bytes()
	return hexutil.Encode(fbytes[:])
}

// HexEncodeToString encodes a byte array into a hex string
func HexEncodeToString(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

// HexDecodeString decodes a hex string into a byte array or error if it failed doing the conversion.
func HexDecodeString(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	return hex.DecodeString(s)
}

// KeccakHash computes the keccak of a stream of bytes. Returns the hex string.
func KeccakHash(stream []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(stream)
	return h.Sum(nil)
}

// HexHashHex parses one or more hex strings into a byte array, hashes it and returns the result
// as a hexstring. If several hex strings are passed, what is hashed is the
// concatenation of the strings and the hasher is implicitly updated only once.
// The hash function is Keccak.
func HexHashHex(v ...string) string {
	buffer := bytes.Buffer{}
	for i := range v {
		decoded, err := hexutil.Decode(v[i])
		if err != nil {
			logrus.Errorf("could not decode `%v` from list `%v`, because `%v`. This can happen when"+
				" the state-manager option is activated but no zk-merkleProof were found", v[i], v, err)
		}
		buffer.Write(decoded)
	}
	res := crypto.Keccak256(buffer.Bytes())
	return hexutil.Encode(res)
}

// HexConcat concatenates hex strings.
func HexConcat(v ...string) string {
	buffer := bytes.Buffer{}
	for i := range v {
		decoded := hexutil.MustDecode(v[i])
		buffer.Write(decoded)
	}
	return hexutil.Encode(buffer.Bytes())
}

// AsBigEndian32Bytes formats an integer as a big-endian uint256.
func AsBigEndian32Bytes(x int) (res [32]byte) {
	new(big.Int).SetInt64(int64(x)).FillBytes(res[:])
	return res
}
