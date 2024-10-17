package serialization

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// WAssignment is an alias for the mapping type used to represent the assignment
// of a column in a [wizard.ProverRuntime]
type WAssignment = collection.Mapping[ifaces.ColID, smartvectors.SmartVector]

// SerializeAssignment serializes map representing the column assignment of a
// wizard protocol.
func SerializeAssignment(a WAssignment) []byte {

	var (
		as    = a.InnerMap()
		ser   = map[string]*CompressedSmartVector{}
		names = a.ListAllKeys()
		lock  = &sync.Mutex{}
	)

	parallel.Execute(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := CompressSmartVector(as[names[i]])
			lock.Lock()
			ser[string(names[i])] = v
			lock.Unlock()
		}
	})

	return serializeAnyWithCborPkg(ser)
}

// DeserializeAssignment deserialize a blob of bytes into a set of column
// assignment representing assigned columns of a Wizard protocol.
func DeserializeAssignment(data []byte) (WAssignment, error) {

	var (
		ser  = map[string]*CompressedSmartVector{}
		res  = WAssignment{}
		lock = &sync.Mutex{}
	)

	if err := deserializeAnyWithCborPkg(data, &ser); err != nil {
		return WAssignment{}, err
	}

	names := make([]string, 0, len(ser))
	for n := range ser {
		names = append(names, n)
	}

	parallel.Execute(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := ser[names[i]].Decompress()
			lock.Lock()
			res.InsertNew(ifaces.ColID(names[i]), v)
			lock.Unlock()
		}
	})

	return res, nil
}

// CompressedSmartVector represents a [smartvectors.SmartVector] in a more
// space-efficient manner.
type CompressedSmartVector struct {
	F []CompressedSVFragment
}

// CompressedSVFragment represent a portion of a SerializableSmartVector
type CompressedSVFragment struct {
	// L is the byte length used by the fragment
	L uint8
	// X is the value used to represent a single field element
	X *big.Int
	// V is a byteslice storing the bytes of a vector if the fragment represent
	// plain values.
	V []byte
	// N is the number of repetion used
	N int
}

func CompressSmartVector(sv smartvectors.SmartVector) *CompressedSmartVector {

	switch v := sv.(type) {
	case *smartvectors.Constant:
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newConstantSVFragment(v.Val(), v.Len()),
			},
		}
	case *smartvectors.Regular:
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newSliceSVFragment(*v),
			},
		}
	case *smartvectors.PaddedCircularWindow:
		var (
			w          = v.Window()
			offset     = v.Offset()
			paddingVal = v.PaddingVal()
			fullLen    = v.Len()
		)

		// It's a left-padded value
		if offset == 0 {
			return &CompressedSmartVector{
				F: []CompressedSVFragment{
					newSliceSVFragment(w),
					newConstantSVFragment(paddingVal, fullLen-len(w)),
				},
			}
		}

		// It's a right-padded value
		if offset+len(w) == fullLen {
			return &CompressedSmartVector{
				F: []CompressedSVFragment{
					newConstantSVFragment(paddingVal, fullLen-len(w)),
					newSliceSVFragment(w),
				},
			}
		}

	}

	// The other cases are not expected, we still support them via a
	// suboptimal method.
	return &CompressedSmartVector{
		F: []CompressedSVFragment{
			newSliceSVFragment(sv.IntoRegVecSaveAlloc()),
		},
	}
}

func (sv *CompressedSmartVector) Decompress() smartvectors.SmartVector {

	if len(sv.F) == 1 && sv.F[0].isConstant() {
		val := new(field.Element).SetBigInt(sv.F[0].X)
		return smartvectors.NewConstant(*val, sv.F[0].N)
	}

	if len(sv.F) == 1 && sv.F[0].isPlain() {
		return smartvectors.NewRegular(sv.F[0].readSlice())
	}

	if len(sv.F) == 2 && sv.F[0].isConstant() && sv.F[1].isPlain() {

		var (
			paddingVal = new(field.Element).SetBigInt(sv.F[0].X)
			window     = sv.F[1].readSlice()
			size       = sv.F[0].N + len(window)
		)

		return smartvectors.LeftPadded(window, *paddingVal, size)
	}

	if len(sv.F) == 2 && sv.F[1].isConstant() && sv.F[0].isPlain() {

		var (
			paddingVal = new(field.Element).SetBigInt(sv.F[1].X)
			window     = sv.F[0].readSlice()
			size       = sv.F[1].N + len(window)
		)

		return smartvectors.RightPadded(window, *paddingVal, size)
	}

	panic("unexpected pattern")
}

func (f *CompressedSVFragment) isConstant() bool {
	return f.X == nil
}

func (f *CompressedSVFragment) isPlain() bool {
	return f.V != nil
}

func (f *CompressedSVFragment) readSlice() []field.Element {

	var (
		l   = f.L
		buf = bytes.NewBuffer(f.V)
		tmp = [32]byte{}
		n   = len(f.V) / 8
		res = make([]field.Element, n)
	)

	for i := range res {
		buf.Read(tmp[32-l:])
		res[i].SetBytes(tmp[:])
	}

	return res
}

func newConstantSVFragment(x field.Element, n int) CompressedSVFragment {

	var (
		f big.Int
		_ = x.BigInt(&f)
	)

	return CompressedSVFragment{
		X: &f,
		N: n,
	}
}

func newSliceSVFragment(fv []field.Element) CompressedSVFragment {

	var (
		l int
	)

	for i := range fv {
		l = max(l, (fv[i].BitLen()+7)/8)
	}

	var (
		res    = make([]byte, 0, len(fv)*l)
		resBuf = bytes.NewBuffer(res)
	)

	for i := range fv {
		fbytes := fv[i].Bytes()
		resBuf.Write(fbytes[32-l:])
	}

	return CompressedSVFragment{
		L: uint8(l),
		V: resBuf.Bytes(),
	}
}
