package v0

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// A Header is a list of batches of blocks of len(blocks)
// len(Header) == nb of batches in the blob
// len(Header[i]) == nb of blocks in the batch i
// Header[i][j] == len (bytes) of the j-th block in the batch i
type Header struct {
	DictChecksum [fr.Bytes]byte
	table        [][]int
	// @alex: temporary needed until we have header-parsing in-circuit
	byteSizeUnpacked int
}

func (s *Header) Equals(other *Header) bool {
	return s.CheckEquality(other) == nil
}

// CheckEquality similar to Equals but returning a description of the mismatch,
// returning nil if the objects are equal
func (s *Header) CheckEquality(other *Header) error {
	if other == nil {
		return errors.New("empty header")
	}
	if s.DictChecksum != other.DictChecksum {
		return errors.New("dictionary mismatch")
	}

	// we ignore batches of len(0), since caller could have
	// call StartNewBatch() without adding any block
	small, large := s, other
	if len(s.table) > len(other.table) {
		small, large = other, s
	}

	absJ := 0
	for i := range small.table {
		if len(small.table[i]) != len(large.table[i]) {
			return fmt.Errorf("batch size mismatch at #%d", i)
		}
		for j := range small.table[i] {
			if small.table[i][j] != large.table[i][j] {
				return fmt.Errorf("block size mismatch at block #%d of batch #%d, #%d total", j, i, absJ+j)
			}
		}
		absJ += len(small.table[i])
	}

	// remaining batches of large should be empty
	for i := len(small.table); i < len(large.table); i++ {
		if len(large.table[i]) != 0 {
			return errors.New("batch count mismatch")
		}
	}

	return nil
}

func (s *Header) NbBatches() int {
	return len(s.table)
}

func (s *Header) NbBlocksInBatch(i int) int {
	return len(s.table[i])
}

func (s *Header) BlockLength(i, j int) int {
	return s.table[i][j]
}

func (s *Header) addBatch() {
	s.table = append(s.table, make([]int, 0, 128))
}

func (s *Header) ByteSize() int {
	return s.byteSizeUnpacked
}

func (s *Header) ByteSizePacked() int {
	return s.byteSizeUnpacked + utils.DivCeil(s.byteSizeUnpacked*8, packingSizeU256)
}

// addBlock adds a block to the last batch
func (s *Header) addBlock(blockLength int) {
	batchID := len(s.table) - 1
	s.table[batchID] = append(s.table[batchID], blockLength)
}

func (s *Header) removeLastBlock() {
	batchID := len(s.table) - 1
	s.table[batchID] = s.table[batchID][:len(s.table[batchID])-1]
}

func (s *Header) resetTable() {
	s.table = s.table[:0]
}

// WriteTo writes the header table to w.
func (s *Header) WriteTo(w io.Writer) (int64, error) {
	var written int64

	// write dictChecksum (32 bytes)
	if _, err := w.Write(s.DictChecksum[:]); err != nil {
		return written, err
	}
	written += 32

	// write nbBatches (uint16)
	nbBatches := uint16(len(s.table))

	if err := binary.Write(w, binary.LittleEndian, nbBatches); err != nil {
		return written, err
	}
	written += 2

	for _, batch := range s.table {
		// write nbBlocksInBatch (uint16)
		nbBlocksInBatch := uint16(len(batch))
		if int(nbBlocksInBatch) != len(batch) {
			return written, errors.New("nb blocks in batch too big: bigger than uint16")
		}

		if err := binary.Write(w, binary.LittleEndian, nbBlocksInBatch); err != nil {
			return written, err
		}
		written += 2

		// write blockLength (uint24) for each block
		for _, blockLength := range batch {
			const maxUint24 = 1<<24 - 1
			if blockLength > maxUint24 {
				return written, errors.New("block length too big: bigger than uint24")
			}

			// write the blockLength on 3 bytes as a uint24 (LittleEndian)
			if _, err := w.Write([]byte{byte(blockLength), byte(blockLength >> 8), byte(blockLength >> 16)}); err != nil {
				return written, err
			}
			written += 3
		}
	}
	return written, nil
}

// ReadFrom reads the header table from r.
func (s *Header) ReadFrom(r io.Reader) (int64, error) {
	var read int64

	// read dictChecksum (32 bytes)
	if _, err := io.ReadFull(r, s.DictChecksum[:]); err != nil {
		return read, err
	}
	read += 32

	// read nbBatches (uint16)
	var nbBatches uint16

	if err := binary.Read(r, binary.LittleEndian, &nbBatches); err != nil {
		return read, err
	}
	read += 2

	s.table = make([][]int, nbBatches)

	var buf [3]byte

	for i := uint16(0); i < nbBatches; i++ {
		// read nbBlocksInBatch (uint16)
		var nbBlocksInBatch uint16
		if err := binary.Read(r, binary.LittleEndian, &nbBlocksInBatch); err != nil {
			return read, err
		}
		read += 2

		s.table[i] = make([]int, nbBlocksInBatch)
		for j := uint16(0); j < nbBlocksInBatch; j++ {
			// read blockLength (uint24) for each block
			var blockLength uint32

			// read the blockLength on 3 bytes as a uint24 (LittleEndian)
			if _, err := io.ReadFull(r, buf[:]); err != nil {
				return read, err
			}
			blockLength = uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16
			read += 3

			s.table[i][j] = int(blockLength)
		}
	}

	s.byteSizeUnpacked = int(read)

	return read, nil
}
