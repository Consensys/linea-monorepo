package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
)

// EthAddress represents an ethereum address. It consists of 20bytes and is the
// obtained by hashing the corresponding ECDSA-secp256k1 public key and keeping
// the 20 first bytes.
type EthAddress [20]byte

// Marshal "e" into JSON format
func (e EthAddress) MarshalJSON() ([]byte, error) {
	hexAddress := utils.HexEncodeToString(e[:])
	marshalled, err := json.Marshal(hexAddress)
	if err != nil {
		return nil, fmt.Errorf("could not marshal eth address : %w", err)
	}
	return marshalled, nil
}

// Unmarshal an ethereum address from JSON format. The expected format is an hex
// string of 20 bytes. If the input string is larger than 20 bytes, truncate the
// leftmost bytes of the input address. If it is too small zero pad it on the
// left.
func (e *EthAddress) UnmarshalJSON(b []byte) error {
	decoded, err := DecodeQuotedHexString(b)
	if err != nil {
		return fmt.Errorf("could not decode eth address %v : %w", string(b), err)
	}

	switch {
	case len(decoded) < 20:
		copy((*e)[20-len(decoded):], decoded)

	case len(decoded) >= 20:
		copy((*e)[:], decoded[len(decoded)-20:])
	}

	return nil
}

// Writes the bytes32 into the given write.
func (e EthAddress) WriteTo(w io.Writer) (int64, error) {
	_, err := w.Write(e[:])
	if err != nil {
		panic(err) // hard forbid any error
	}
	return 20, nil
}

// Reads a bytes32 from the given reader
func (e *EthAddress) ReadFrom(r io.Reader) (int64, error) {
	n, err := r.Read((*e)[:])
	return int64(n), err
}

// Reads an Ethereum address that is left zero-padded to 32 bytes. Example:
// 0x000000000000000000000000deadbeefdeadbeefdeadbeefdeadbeefdeadbeef
func (e *EthAddress) ReadFrom32BytesLeftZeroPadded(r io.Reader) (int64, error) {
	var buf [12]byte
	n, err := r.Read(buf[:])
	if err != nil {
		return int64(n), fmt.Errorf("could not read 32 bytes, left-zero padded ethereum address: %v", err)
	}
	return e.ReadFrom(r)
}

// Writes an Ethereum address on 32 using left-zero padding.
func (e *EthAddress) WriteOn32Bytes(w io.Writer) (int64, error) {
	buf := [12]byte{}
	_, err0 := w.Write(buf[:])
	_, err1 := e.WriteTo(w)
	err := errors.Join(err0, err1)
	if err != nil {
		return 0, fmt.Errorf("writing address %x on 32 bytes failed: %w", *e, err)
	}
	return 32, nil
}

func (e EthAddress) Hex() string {
	return utils.HexEncodeToString(e[:])
}

func AddressFromHex(h string) (EthAddress, error) {
	byt, err := utils.HexDecodeString(h)
	if err != nil {
		return EthAddress{}, fmt.Errorf("reading hex address %s : %w", h, err)
	}
	if len(byt) != 20 {
		return EthAddress{}, fmt.Errorf("reading hex address %s : has %d bytes", h, len(byt))
	}
	var res EthAddress
	copy(res[:], byt)
	return res, nil
}

// Construct a dummy address from an integer
func DummyAddress(i int) (a EthAddress) {
	a[0] = byte(i)
	return a
}
