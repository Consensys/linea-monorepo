package lzss

import (
	"encoding/binary"
	"io"
)

const (
	// Version is the current release version of the compressor.
	Version    = 0
	HeaderSize = 3
)

// Header is the header of a compressed data.
// It contains the compressor release version and the compression level.
type Header struct {
	Version uint16 // compressor release version
	Level   Level  // compression level
	// A future version may add more fields here, so we'll have to read the
	// version first, then read the rest of the header based on the version.
	// Extra   []byte // "extra data", max len == math.MaxUint16
}

func (s *Header) WriteTo(w io.Writer) (int64, error) {
	if err := binary.Write(w, binary.LittleEndian, uint16(s.Version)); err != nil {
		return 0, err
	}

	if _, err := w.Write([]byte{byte(s.Level)}); err != nil {
		return 2, err
	}

	return HeaderSize, nil
}

func (s *Header) ReadFrom(r io.Reader) (int64, error) {
	var b [HeaderSize]byte
	n, err := io.ReadFull(r, b[:])
	if err != nil {
		return int64(n), err
	}

	s.Version = binary.LittleEndian.Uint16(b[:2])
	s.Level = Level(b[2])
	return int64(n), nil
}
