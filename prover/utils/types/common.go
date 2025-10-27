package types

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/utils"
)

func DecodeQuotedHexString(b []byte) ([]byte, error) {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return nil, fmt.Errorf(
			"could not unmarshal hex string : expected a quoted string but got `%v`, error : %w",
			string(b), err,
		)
	}
	decoded, err := utils.HexDecodeString(unquoted)
	if err != nil {
		return nil, fmt.Errorf(
			"could not unmarshal hex string : expected an hex string but got `%v`, error : %w",
			unquoted, err,
		)
	}
	return decoded, nil
}

func MarshalHexBytesJSON(b []byte) []byte {
	hexstring := utils.HexEncodeToString(b)
	return []byte(strconv.Quote(hexstring))
}

// LeftPadded copies b into a new byte slice of size 2*len(b).
func LeftPadded(b []byte) []byte {

	if len(b)%2 != 0 {
		panic("input length must be even")
	}

	res := make([]byte, len(b)*2)

	// We copy every 2 bytes from tmp into res, left started by 2 zero-bytes.
	for i := 0; i < len(b)/2; i++ {
		copy(res[4*i+2:4*i+4], b[2*i:2*i+2])
	}
	return res
}

// RemovePadding removes the zero padding from the byte slice.
func RemovePadding(b []byte) []byte {

	if len(b)%4 != 0 {
		panic("input length must be a multiple of 4")
	}

	data := make([]byte, len(b)/2)
	for i := 0; i < len(b)/4; i++ {
		copy(data[2*i:2*i+2], b[4*i+2:4*i+4])
	}
	return data
}

func WriteInt64On16Bytes(w io.Writer, x int64) (int64, error) {
	xBytes := [8]byte{}

	// Convert the int64 to its 8-byte representation
	binary.BigEndian.PutUint64(xBytes[:], uint64(x))

	// We copy every 2 bytes from tmp into res, left padded by 2 zero-bytes.
	res := LeftPadded(xBytes[:])

	n, err := w.Write(res[:])
	if err != nil {
		return int64(n), fmt.Errorf("could not write 16 bytes into Writer : %w", err)
	}
	return int64(n), nil
}

func ReadInt64On16Bytes(r io.Reader) (x, n_ int64, err error) {
	var buf [16]byte

	// Read exactly 16 bytes. io.ReadFull handles partial reads and EOF
	// correctly, returning io.ErrUnexpectedEOF if 16 bytes aren't available.
	n, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, int64(n), fmt.Errorf("could not read 16 bytes: %w", err)
	}

	// De-interleave the data from the 16-byte buffer into an 8-byte buffer
	data := RemovePadding(buf[:])

	// Convert the 8 data bytes back to a uint64
	xU64 := binary.BigEndian.Uint64(data[:])

	if n < 0 {
		panic("we are only reading 8 bits so this should not overflow")
	}
	xU64 &= 0x7fffffffffffffff  // TODO delete this if negative numbers are allowed
	return int64(xU64), 16, err // #nosec G115 -- above line precludes overflowing
}

// Big int are assumed to fit on 64 bytes and are written as a single
// block of 64 bytes in bigendian form (i.e, zero-padded on the left)
func WriteBigIntOn64Bytes(w io.Writer, b *big.Int) (int64, error) {

	balanceBig := b.FillBytes(make([]byte, 32))
	balanceBigPadded := LeftPadded(balanceBig)
	n, err := w.Write(balanceBigPadded)
	return int64(n), err
}

// Read a bigint from a reader assuming it takes 32 bytes
func ReadBigIntOn64Bytes(r io.Reader) (*big.Int, error) {
	buf := [64]byte{}
	_, err := r.Read(buf[:])
	res := RemovePadding(buf[:])
	if err != nil {
		return nil, fmt.Errorf("reading big int, could not 32 bytes from reader: %v", err)
	}
	bi := new(big.Int).SetBytes(res[:])
	if bi.IsUint64() && bi.Uint64() == 0 {
		// there is an edge-case here that breaks the test. Namely, giving 32
		// zero bytes to a big.Int or calling NewInt with `0` do not produce the
		// same instances although they are both equivalent. This case handling
		// aims at preventing this.
		bi = big.NewInt(0)
	}

	return bi, nil
}
