package lzss

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

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

func CompressedStreamInfo(c, dict []byte) (CompressionPhrases, error) {
	in := bitio.NewReader(bytes.NewReader(c))

	// parse header
	var header Header
	sizeHeader, err := header.ReadFrom(in)
	if err != nil {
		return nil, err
	}
	if header.Version != Version {
		panic("unsupported compressor version")
	}
	if header.Level == NoCompression {
		return CompressionPhrases{{
			Type:              0,
			Length:            len(c) - int(sizeHeader),
			ReferenceAddress:  0,
			StartDecompressed: 0,
			StartCompressed:   0,
			Content:           c[sizeHeader:],
		}}, nil
	}

	var res CompressionPhrases

	// init dict and backref types
	dict = AugmentDict(dict)
	shortBackRefType, longBackRefType, dictBackRefType := InitBackRefTypes(len(dict), header.Level)

	bDict := backref{bType: dictBackRefType}
	bShort := backref{bType: shortBackRefType}
	bLong := backref{bType: longBackRefType}

	var out bytes.Buffer
	out.Grow(len(c) * 7)

	// the decompressor considers the direct copying of each byte of the input its own event.
	// that's inconvenient to the human eye, so we group all consecutive literal copies into the same event
	// literalCopyStart is the index of the first byte of the literal copy in the DECOMPRESSED stream.
	// it is -1 if we are not currently copying a literal
	literalCopyStart := -1
	inI := 0

	emitLiteralIfNecessary := func() {
		if literalCopyStart == -1 {
			return
		}
		res = append(res, CompressionPhrase{
			Type:              0,
			Length:            out.Len() - literalCopyStart,
			ReferenceAddress:  literalCopyStart,
			StartDecompressed: literalCopyStart,
			StartCompressed:   inI,
			Content:           out.Bytes()[literalCopyStart:],
		})
		inI += (out.Len() - literalCopyStart) * 8
		literalCopyStart = -1
	}

	emitRef := func(b *backref) {
		addr := out.Len() - b.length - b.address // this happens post writing out the backref
		if b.bType == dictBackRefType {
			addr = b.address
		}
		res = append(res, CompressionPhrase{
			Type:              b.bType.Delimiter,
			Length:            b.length,
			ReferenceAddress:  addr,
			StartDecompressed: out.Len() - b.length,
			StartCompressed:   inI,
			Content:           out.Bytes()[out.Len()-b.length:],
		})
		inI += int(b.bType.NbBitsBackRef)
	}

	// read byte per byte; if it's a backref, write the corresponding bytes
	// otherwise, write the byte as is
	s := in.TryReadByte()
	for in.TryError == nil {
		switch s {
		case SymbolShort:
			emitLiteralIfNecessary()
			// short back ref
			if err := bShort.readFrom(in); err != nil {
				return nil, err
			}
			for i := 0; i < bShort.length; i++ {
				out.WriteByte(out.Bytes()[out.Len()-bShort.address])
			}
			emitRef(&bShort)
		case SymbolLong:
			emitLiteralIfNecessary()
			// long back ref
			if err := bLong.readFrom(in); err != nil {
				return nil, err
			}
			for i := 0; i < bLong.length; i++ {
				out.WriteByte(out.Bytes()[out.Len()-bLong.address])
			}
			emitRef(&bLong)
		case SymbolDict:
			emitLiteralIfNecessary()
			// dict back ref
			if err := bDict.readFrom(in); err != nil {
				return nil, err
			}
			out.Write(dict[bDict.address : bDict.address+bDict.length])
			emitRef(&bDict)
		default:
			if literalCopyStart == -1 {
				literalCopyStart = out.Len()
			}
			out.WriteByte(s)
		}
		s = in.TryReadByte()
	}
	return res, nil
}

func (c CompressionPhrases) ToCSV() []byte {
	var b bytes.Buffer
	b.WriteString("type,length,start_decompressed (bytes),start_compressed (bits),reference_address,content (hex)\n")
	for _, phrase := range c {
		switch phrase.Type {
		case SymbolShort:
			b.WriteString("short,")
		case SymbolLong:
			b.WriteString("long,")
		case SymbolDict:
			b.WriteString("dict,")
		case 0:
			b.WriteString("literal,")
		default:
			panic("unknown phrase type")
		}

		b.WriteString(strconv.Itoa(phrase.Length))
		b.WriteString(",")

		b.WriteString(strconv.Itoa(phrase.StartDecompressed))
		b.WriteString(",")
		b.WriteString(strconv.Itoa(phrase.StartCompressed))
		b.WriteString(",")
		b.WriteString(strconv.Itoa(phrase.ReferenceAddress))
		b.WriteString(",")
		b.WriteString(hex.EncodeToString(phrase.Content))
		b.WriteString("\n")
	}
	return b.Bytes()
}
