// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vectorext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Vector represents a slice of Element.
//
// It implements the following interfaces:
//   - Stringer
//   - io.WriterTo
//   - io.ReaderFrom
//   - encoding.BinaryMarshaler
//   - encoding.BinaryUnmarshaler
//   - sort.Interface
type Vector []fext.Element

// MarshalBinary implements encoding.BinaryMarshaler
func (vector *Vector) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	if _, err = vector.WriteTo(&buf); err != nil {
		return
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (vector *Vector) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	_, err := vector.ReadFrom(r)
	return err
}

// WriteTo implements io.WriterTo and writes a vector of big endian encoded Element.
// Length of the vector is encoded as a uint32 on the first 4 bytes.
func (vector *Vector) WriteTo(w io.Writer) (int64, error) {

	// encode slice length
	if err := binary.Write(w, binary.BigEndian, uint32(len(*vector))); err != nil {
		return 0, err
	}

	n := int64(4)

	// TODO put that in a const somewhere
	degreeExtension := 4

	bufCoords := make([]koalabear.Element, degreeExtension)
	var buf [koalabear.Bytes]byte

	for i := 0; i < len(*vector); i++ {

		bufCoords[0] = (*vector)[i].B0.A0
		bufCoords[1] = (*vector)[i].B0.A1
		bufCoords[2] = (*vector)[i].B1.A0
		bufCoords[3] = (*vector)[i].B1.A1

		for j := 0; j < degreeExtension; j++ {

			koalabear.BigEndian.PutElement(&buf, bufCoords[j])
			m, err := w.Write(buf[:])
			n += int64(m)
			if err != nil {
				return n, err
			}

		}

	}
	return n, nil
}

// AsyncReadFrom reads a vector of big endian encoded Element.
// Length of the vector must be encoded as a uint32 on the first 4 bytes.
// It consumes the needed bytes from the reader and returns the number of bytes read and an error if any.
// It also returns a channel that will be closed when the validation is done.
// The validation consist of checking that the elements are smaller than the modulus, and
// converting them to montgomery form.
func (vector *Vector) AsyncReadFrom(r io.Reader) (int64, error, chan error) {

	chErr := make(chan error, 1)
	var bufSizeSlice [4]byte
	if read, err := io.ReadFull(r, bufSizeSlice[:]); err != nil {
		close(chErr)
		return int64(read), err, chErr
	}
	sliceLen := binary.BigEndian.Uint32(bufSizeSlice[:])

	n := int64(4)
	(*vector) = make(Vector, sliceLen)
	if sliceLen == 0 {
		close(chErr)
		return n, nil, chErr
	}

	// TODO declare those as a const somewhere
	degreeExtension := 4
	modulus := uint32(2130706433)
	nbBytesE4Elmt := degreeExtension * koalabear.Bytes

	bSlice := unsafe.Slice((*byte)(unsafe.Pointer(&(*vector)[0])), sliceLen*uint32(nbBytesE4Elmt))
	read, err := io.ReadFull(r, bSlice)
	n += int64(read)
	if err != nil {
		close(chErr)
		return n, err, chErr
	}

	var tmp uint32

	go func() {
		var cptErrors uint64

		// process the elements in parallel
		parallel.Execute(int(sliceLen), func(start, end int) {

			coeffs := make([]koalabear.Element, degreeExtension)

			for i := start; i < end; i++ {

				bstart := i * nbBytesE4Elmt

				// read the 4 coordinates an E4 element
				for j := 0; j < degreeExtension; j++ {
					b := bSlice[bstart : bstart+koalabear.Bytes]
					tmp = binary.BigEndian.Uint32(b[:])
					if !(tmp < modulus) {
						atomic.AddUint64(&cptErrors, 1)
						return
					}
					coeffs[j].SetBytes(b) // <- already converted in Montgomery form
					bstart += koalabear.Bytes
				}
				(*vector)[i].B0.A0.Set(&coeffs[0])
				(*vector)[i].B0.A1.Set(&coeffs[1])
				(*vector)[i].B1.A0.Set(&coeffs[2])
				(*vector)[i].B1.A1.Set(&coeffs[3])

			}
		})

		if cptErrors > 0 {
			chErr <- fmt.Errorf("async read: %d elements failed validation", cptErrors)
		}
		close(chErr)
	}()
	return n, nil, chErr
}

// ReadFrom implements io.ReaderFrom and reads a vector of big endian encoded Element.
// Length of the vector must be encoded as a uint32 on the first 4 bytes.
func (vector *Vector) ReadFrom(r io.Reader) (int64, error) {

	var buf [koalabear.Bytes]byte
	var bufSize [4]byte

	if read, err := io.ReadFull(r, bufSize[:]); err != nil {
		return int64(read), err
	}
	sliceLen := binary.BigEndian.Uint32(bufSize[:4])

	n := int64(4)
	(*vector) = make(Vector, sliceLen)

	// TODO should declare the degree of the extension as a const or something
	degreeExtension := 4
	coeffs := make([]koalabear.Element, degreeExtension)
	for i := 0; i < int(sliceLen); i++ {

		for j := 0; j < degreeExtension; j++ {
			read, err := io.ReadFull(r, buf[:])
			n += int64(read)
			if err != nil {
				return n, err
			}
			coeffs[j], err = koalabear.BigEndian.Element(&buf)
			if err != nil {
				return n, err
			}
		}

		(*vector)[i].B0.A0.Set(&coeffs[0])
		(*vector)[i].B0.A1.Set(&coeffs[1])
		(*vector)[i].B1.A0.Set(&coeffs[2])
		(*vector)[i].B1.A1.Set(&coeffs[3])
	}

	return n, nil
}

// String implements fmt.Stringer interface
func (vector Vector) String() string {
	var sbb strings.Builder
	sbb.WriteByte('[')
	for i := 0; i < len(vector); i++ {
		sbb.WriteString(vector[i].String())
		if i != len(vector)-1 {
			sbb.WriteByte(',')
		}
	}
	sbb.WriteByte(']')
	return sbb.String()
}

// Len is the number of elements in the collection.
func (vector Vector) Len() int {
	return len(vector)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (vector Vector) Less(i, j int) bool {
	return vector[i].Cmp(&vector[j]) == -1
}

// Swap swaps the elements with indexes i and j.
func (vector Vector) Swap(i, j int) {
	vector[i], vector[j] = vector[j], vector[i]
}

func innerProductVecByElement(res *fext.Element, a Vector, b []field.Element) {
	if len(a) != len(b) {
		panic("vector.InnerProduct: vectors don't have the same length")
	}
	var tmp fext.Element
	for i := 0; i < len(a); i++ {
		tmp.MulByElement(&a[i], &b[i])
		res.Add(res, &tmp)
	}
}
func PrettifyGeneric(a []fext.GenericFieldElem) string {
	res := "["

	for i := range a {
		// Discards the case first element when adding a comma
		if i > 0 {
			res += ", "
		}

		res += fmt.Sprintf("%v", a[i].String())
	}
	res += "]"

	return res
}
