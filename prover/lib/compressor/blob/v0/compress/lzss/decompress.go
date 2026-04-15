package lzss

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/icza/bitio"
)

// Decompress decompresses the given data using the given dictionary
// the dictionary must be the same as the one used to compress the data
// Note that this is not a fail-safe decompressor, it will fail ungracefully if the data
// has a different format than the one expected
func Decompress(data, dict []byte) (d []byte, err error) {
	in := bitio.NewReader(bytes.NewReader(data))

	// parse header
	var header Header
	sizeHeader, err := header.ReadFrom(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	if header.Version != Version {
		return nil, errors.New("unsupported compressor version")
	}
	if header.Level == NoCompression {
		return data[sizeHeader:], nil
	}

	// init dict and backref types
	dict = AugmentDict(dict)
	shortBackRefType, longBackRefType, dictBackRefType := InitBackRefTypes(len(dict), header.Level)

	bDict := backref{bType: dictBackRefType}
	bShort := backref{bType: shortBackRefType}
	bLong := backref{bType: longBackRefType}

	var out bytes.Buffer
	out.Grow(len(data) * 7)

	// read byte per byte; if it's a backref, write the corresponding bytes
	// otherwise, write the byte as is
	s := in.TryReadByte()
	for in.TryError == nil {
		switch s {
		case SymbolShort:
			// short back ref
			if err := bShort.readFrom(in); err != nil {
				return nil, err
			}
			for i := 0; i < bShort.length; i++ {
				if bShort.address > out.Len() {
					return nil, fmt.Errorf("invalid short backref %v - output buffer is only %d bytes long", bShort, out.Len())
				}
				out.WriteByte(out.Bytes()[out.Len()-bShort.address])
			}
		case SymbolLong:
			// long back ref
			if err := bLong.readFrom(in); err != nil {
				return nil, err
			}
			for i := 0; i < bLong.length; i++ {
				if bLong.address > out.Len() {
					return nil, fmt.Errorf("invalid long backref %v - output buffer is only %d bytes long", bLong, out.Len())
				}
				out.WriteByte(out.Bytes()[out.Len()-bLong.address])
			}
		case SymbolDict:
			// dict back ref
			if err := bDict.readFrom(in); err != nil {
				return nil, err
			}
			if bDict.address > len(dict) || bDict.address+bDict.length > len(dict) {
				return nil, fmt.Errorf("invalid dict backref %v - dict is only %d bytes long", bDict, len(dict))
			}
			out.Write(dict[bDict.address : bDict.address+bDict.length])
		default:
			out.WriteByte(s)
		}
		s = in.TryReadByte()
	}

	return out.Bytes(), nil
}

type CompressionPhrase struct {
	Type              byte
	Length            int
	ReferenceAddress  int
	StartDecompressed int
	StartCompressed   int
	Content           []byte
}

type CompressionPhrases []CompressionPhrase
