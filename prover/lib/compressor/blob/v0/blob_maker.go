package v0

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0/compress/lzss"
	"github.com/icza/bitio"
	"io"
)

const (
	maxOutputSize      = 1 << 20 // 1MB
	packingSizeU256    = 248     // = 31 bytes
	packingSizeLastU64 = 64 - 256 + packingSizeU256

	// These also impact the circuit constraints (compile / setup time)
	MaxUncompressedBytes = 800_000   // defines the max size we can handle for a blob (uncompressed) input
	MaxUsableBytes       = 32 * 4096 // TODO @gbotrel confirm this value // defines the number of bytes available in a blob
)

// DecompressBlob decompresses a blob and returns the header and the blocks as they were compressed.
// rawBlocks is the raw payload of the blob, delivered in packed format @TODO bad idea. fix
func DecompressBlob(b []byte, dictStore dictionary.Store) (blobHeader *Header, rawBlocks []byte, blocks [][]byte, err error) {
	// UnpackAlign the blob
	b, err = UnpackAlign(b)
	if err != nil {
		return nil, nil, nil, err
	}

	// read the header
	blobHeader = new(Header)
	read, err := blobHeader.ReadFrom(bytes.NewReader(b))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read blob header: %w", err)
	}

	// retrieve dict
	dict, err := dictStore.Get(blobHeader.DictChecksum[:], 0)
	if err != nil {
		return nil, nil, nil, err
	}

	b = b[read:]

	// decompress the data
	rawBlocks, err = lzss.Decompress(b, dict)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decompress blob body: %w", err)
	}

	offset := 0
	for _, batch := range blobHeader.table {
		for _, blockLength := range batch {
			blocks = append(blocks, rawBlocks[offset:offset+blockLength])
			offset += blockLength
		}
	}

	buf := &bytes.Buffer{}
	if _, err := PackAlign(buf, rawBlocks, nil); err != nil {
		return nil, nil, nil, fmt.Errorf("error packing raw blocks: %w", err)
	}

	return blobHeader, buf.Bytes(), blocks, nil
}

// PackAlignSize returns the size of the data when packed with PackAlign.
func PackAlignSize(a, b []byte) (n int) {
	nbBytes := len(a) + len(b) + 1 // + 1 for the padding length
	nbBits := nbBytes * 8

	// we may need to add some bits to a and b to ensure we can process some blocks of 248 bits
	extraBits := (packingSizeU256 - nbBits%packingSizeU256) % packingSizeU256
	nbBits += extraBits

	return ((nbBits / packingSizeU256) * 32)
}

// packAlignSizeToRefactor returns the size of the data when packed with PackAlign (but takes len of the slices as input).
// TODO @gbotrel refactor and reconcile with @Tabaie PR
func packAlignSizeToRefactor(lena, lenb int) (n int) {
	nbBytes := lena + lenb + 1 // + 1 for the padding length
	nbBits := nbBytes * 8

	// we may need to add some bits to a and b to ensure we can process some blocks of 248 bits
	extraBits := (packingSizeU256 - nbBits%packingSizeU256) % packingSizeU256
	nbBits += extraBits

	return ((nbBits / packingSizeU256) * 32)
}

// PackAlign writes a and b to w, aligned to fr.Element (bls12-377) boundary.
// It returns the length of the data written to w.
func PackAlign(w io.Writer, a, b []byte) (n int64, err error) {

	nbBytes := len(a) + len(b) + 1 // + 1 for the padding length
	nbBits := nbBytes * 8

	// we may need to add some bits to a and b to ensure we can process some blocks of 248 bits
	extraBits := (packingSizeU256 - nbBits%packingSizeU256) % packingSizeU256
	nbBits += extraBits

	if nbBits%packingSizeU256 != 0 {
		return 0, fmt.Errorf("nbBits mod %d != 0, (nbBits = %d)", packingSizeU256, nbBits)
	}

	// padding will always be less than 31 bytes
	bytePadding := (extraBits + 7) / 8
	var buf [32]byte

	// the last non-zero byte of the stream is the padding length
	// we add +1 to handle the case where the padding length is 0
	buf[0] = uint8(bytePadding + 1)

	r := bitio.NewReader(io.MultiReader(bytes.NewReader(a), bytes.NewReader(b), bytes.NewReader(buf[:bytePadding+1])))

	var tryWriteErr error
	tryWrite := func(v uint64) {
		if tryWriteErr == nil {
			tryWriteErr = binary.Write(w, binary.BigEndian, v)
		}
	}

	for i := 0; i < nbBits/packingSizeU256; i++ {
		tryWrite(r.TryReadBits(packingSizeLastU64))
		tryWrite(r.TryReadBits(64))
		tryWrite(r.TryReadBits(64))
		tryWrite(r.TryReadBits(64))
	}

	if tryWriteErr != nil {
		return 0, fmt.Errorf("when writing to w: %w", tryWriteErr)
	}

	if r.TryError != nil {
		return 0, fmt.Errorf("when reading from multi-reader: %w", r.TryError)
	}

	n1 := ((nbBits / packingSizeU256) * 32)
	if n1 != PackAlignSize(a, b) {
		return 0, errors.New("inconsistent PackAlignSize")
	}
	return int64(n1), nil
}

// UnpackAlign unpacks r (packed with PackAlign) and returns the unpacked data.
func UnpackAlign(r []byte) ([]byte, error) {
	if len(r)%32 != 0 {
		return nil, errors.New("invalid data length; expected multiple of 32")
	}
	n := len(r) / 32
	var out bytes.Buffer
	w := bitio.NewWriter(&out)
	for i := 0; i < n; i++ {
		// read 32bytes
		element := r[32*i : 32*(i+1)]
		// write 248 bits
		w.TryWriteBits(binary.BigEndian.Uint64(element[0:8]), packingSizeLastU64)
		w.TryWriteBits(binary.BigEndian.Uint64(element[8:16]), 64)
		w.TryWriteBits(binary.BigEndian.Uint64(element[16:24]), 64)
		w.TryWriteBits(binary.BigEndian.Uint64(element[24:32]), 64)
	}
	if w.TryError != nil {
		return nil, fmt.Errorf("when writing to bitio.Writer: %w", w.TryError)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("when closing bitio.Writer: %w", err)
	}

	// the last non-zero byte is the padding length + 1
	// this is a cheap sanity check; we could also just resize the output buffer
	// to the correct size.
	cpt := 0
	for out.Bytes()[out.Len()-1] == 0 {
		out.Truncate(out.Len() - 1)
		cpt++
	}
	// last byte should be equal to cpt
	lastNonZero := out.Bytes()[out.Len()-1]
	if (cpt % 31) != int(lastNonZero)-1 {
		return nil, errors.New("invalid padding length")
	}
	out.Truncate(out.Len() - 1)

	return out.Bytes(), nil
}
