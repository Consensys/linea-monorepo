package field

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"reflect"
	"runtime"

	"github.com/consensys/gnark-crypto/field/babybear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

var (
	// RootPowers = []int{1, 3}, v^2=u and u^2=3.
	RootPowers = []int{1, 3}
)

const (
	// ExtensionDegree is the degree of the extension Ext over Fr
	ExtensionDegree int = 4
)

func UintsToExt(v1, v2, v3, v4 uint64) Ext {
	var z Ext
	z.B0.A0.SetUint64(v1)
	z.B0.A1.SetUint64(v2)
	z.B1.A0.SetUint64(v3)
	z.B1.A1.SetUint64(v4)
	return z
}

func UintToExt(b uint64) Ext {
	var res Ext
	res.B0.A0.SetUint64(b)
	return res
}

func IntsToExt(v1, v2, v3, v4 int64) Ext {
	var z Ext
	z.B0.A0.SetInt64(v1)
	z.B0.A1.SetInt64(v2)
	z.B1.A0.SetInt64(v3)
	z.B1.A1.SetInt64(v4)
	return z
}

// NewFromString only sets the first coordinate of the field extension
func StringToExt(s string) (res Ext) {
	res.B0.A0 = field.NewFromString(s)
	return res
}

// One returns 1
func OneExt() Ext {
	var res Ext
	res.B0.A0.SetOne()
	return res
}

// SetZero sets an E4 elmt to zero
func ZeroExt() Ext {
	var res Ext
	return res
}

// IsFr returns if the receiver represents a base field element
func (z *Ext) IsFr() bool {
	return z.B0.A1[0] == 0 && z.B1.A0[0] == 0 && z.B1.A1[0] == 0
}

// TryIntoFr attempts to convert the receiver to a base field element
func (z *Ext) TryIntoFr() (Fr, bool) {
	if z.IsFr() {
		return Fr(z.B0.A0), true
	}
	return Fr{}, false
}

// SetFr sets the value of the receiver to a base field element
func (z *Ext) SetFr(x *Fr) *Ext {
	z.B0.A0 = *unsafeCast[Fr, koalabear.Element](x)
	z.zeroesImaginaryPart()
	return z
}

// AddFr adds a base field and an extension field and set the result in the
// receiver and then returns the receiver pointer.
func (z *Ext) AddFr(x *Ext, y *Fr) *Ext {
	z.B0.A0.Add(&x.B0.A0, unsafeCast[Fr, koalabear.Element](y))
	z.B0.A1 = x.B0.A1
	z.B1 = x.B1
	return z
}

// SubFrFromBase substracts a base field from an extension field and set the result in the
// receiver and then returns the receiver pointer.
func (z *Ext) SubFrFromBase(x *Ext, y *Fr) *Ext {
	z.B0.A0.Sub(&x.B0.A0, unsafeCast[Fr, koalabear.Element](y))
	z.B0.A1 = x.B0.A1
	z.B1 = x.B1
	return z
}

// SubExtFromFr substracts a field extension from a base field and sets the result
// in the receiver and then returns the receiver pointer.
func (z *Ext) SubExtFromFr(x *Fr, y *Ext) *Ext {
	z.B0.A0.Sub(unsafeCast[Fr, koalabear.Element](x), &y.B0.A0)
	z.B0.A1.Neg(&y.B0.A1)
	z.B1.Neg(&y.B1)
	return z
}

// DivExtByFr divides a base field and an extension field and set the result in the
// receiver and then returns the receiver pointer.
func (z *Ext) DivExtByFr(x *Ext, y *Fr) *Ext {
	var y_ Fr
	y_.Inverse(y)
	return z.MulByElement(x, &y_)
}

// DivFrByExt divides a base field and an extension field and set the result in the
// receiver and then returns the receiver pointer.
func (z *Ext) DivFrByExt(x *Fr, y *Ext) *Ext {
	var y_ Ext
	y_.Inverse(y)
	return z.MulByElement(&y_, x)
}

// Text returns a string representation of the field element
func (z *Ext) Text(base int) string {
	if base < 2 || base > 36 {
		panic("invalid base")
	}
	if z == nil {
		return "<nil>"
	}

	res := fmt.Sprintf("%s + %s*u + (%s + %s*u)*v", z.B0.A0.Text(base), z.B0.A1.Text(base), z.B1.A0.Text(base), z.B1.A1.Text(base))
	return res
}

// SetInterface sets a field element from an interface object
func (z *Ext) SetInterface(i1 interface{}) (*Ext, error) {
	if i1 == nil {
		return nil, errors.New("can't set fr.Ext with <nil>")
	}

	switch c1 := i1.(type) {
	case Ext:
		return z.Set(&c1), nil
	case *Ext:
		if c1 == nil {
			return nil, errors.New("can't set fext.Ext with <nil>")
		}
		return z.Set(c1), nil
	case GenericFieldElem:
		return z.Set(&c1.Ext), nil
	case *GenericFieldElem:
		if c1 == nil {
			return nil, errors.New("can't set fext.Ext with <nil>")
		}
		return z.Set(&c1.Ext), nil
	case Fr:
		return z.SetFr(&c1), nil
	case *Fr:
		if c1 == nil {
			return nil, errors.New("can't set fext.Ext with <nil>")
		}
		return z.SetFr(c1), nil
	case uint8:
		return z.SetFromUInt(uint64(c1)), nil
	case uint16:
		return z.SetFromUInt(uint64(c1)), nil
	case uint32:
		return z.SetFromUInt(uint64(c1)), nil
	case uint:
		return z.SetFromUInt(uint64(c1)), nil
	case uint64:
		return z.SetFromUInt(c1), nil
	case int8:
		return z.SetFromInt(int64(c1)), nil
	case int16:
		return z.SetFromInt(int64(c1)), nil
	case int32:
		return z.SetFromInt(int64(c1)), nil
	case int64:
		return z.SetFromInt(c1), nil
	case int:
		return z.SetFromInt(int64(c1)), nil
	case string:
		z.B0.A0.SetString(c1)
		z.zeroesImaginaryPart()
		return z, nil
	case *big.Int:
		if c1 == nil {
			return nil, errors.New("can't set fr.Ext with <nil>")
		}
		z.B0.A0.SetBigInt(c1)
		z.zeroesImaginaryPart()
		return z, nil
	case big.Int:
		z.B0.A0.SetBigInt(&c1)
		z.zeroesImaginaryPart()
		return z, nil
	case []byte:
		z := SetBytes(c1)
		return &z, nil
	default:
		return nil, errors.New("can't set fr.Ext from type " + reflect.TypeOf(i1).String())
	}
}

func (z *Ext) Uint64s() (uint64, uint64, uint64, uint64) {
	return uint64(z.B0.A0.Bits()[0]), uint64(z.B0.A1.Bits()[0]), uint64(z.B1.A0.Bits()[0]), uint64(z.B1.A1.Bits()[0])
}

// SetFromUInt sets z to v and returns z
func (z *Ext) SetFromUInt(v uint64) *Ext {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	z.B0.A0.SetUint64(v)
	z.zeroesImaginaryPart()
	return z // z.toMont()
}

func (z *Ext) SetFromInt(v int64) *Ext {
	z.B0.A0.SetInt64(v)
	z.zeroesImaginaryPart()
	return z // z.toMont()
}

// Lift converts a base field element to an extension field element
func (z *Fr) Lift() Ext {
	var res Ext
	res.B0.A0.Set(unsafeCast[Fr, koalabear.Element](z))
	return res
}

// Lift returns a copy of z. It is only implemented for genericity reasons.
func (z *Ext) Lift() Ext {
	return *z
}

// RandomExt generates a random and non-reproducible field extension element.
func RandomExt() Ext {
	var res Ext
	res.SetRandom()
	return res
}

// PseudoRandExt generates a pseudo-random field extension element.
func PseudoRandExt(rng *rand.Rand) Ext {
	var result Ext
	result.B0.A0 = field.PseudoRand(rng)
	result.B0.A1 = field.PseudoRand(rng)
	result.B1.A0 = field.PseudoRand(rng)
	result.B1.A1 = field.PseudoRand(rng)
	return result
}

func (z *Ext) ExpToInt(x Ext, k int) *Ext {
	if k == 0 {
		return z.SetOne()
	}

	if k < 0 {
		x.Inverse(&x)
		k = -k
	}

	z.Set(&x)

	for i := bits.Len(uint(k)) - 2; i >= 0; i-- {
		z.Square(z)
		if (k>>i)&1 == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

// Bytes returns the value of z as a big-endian byte array the output is 16
// bytes, not Bytes32
func (z *Ext) Bytes() (res [field.Bytes * 4]byte) {
	var result [field.Bytes * 4]byte

	valBytes := z.B0.A0.Bytes()
	copy(result[0:field.Bytes], valBytes[:])

	valBytes = z.B0.A1.Bytes()
	copy(result[field.Bytes:2*field.Bytes], valBytes[:])

	valBytes = z.B1.A0.Bytes()
	copy(result[2*field.Bytes:3*field.Bytes], valBytes[:])

	valBytes = z.B1.A1.Bytes()
	copy(result[3*field.Bytes:4*field.Bytes], valBytes[:])

	return result
}

func SetBytes(data []byte) Ext {
	var res Ext
	res.B0.A0 = koalabear.Element{binary.BigEndian.Uint32(data[0:4])}
	res.B0.A1 = koalabear.Element{binary.BigEndian.Uint32(data[4:8])}
	res.B1.A0 = koalabear.Element{binary.BigEndian.Uint32(data[8:12])}
	res.B1.A1 = koalabear.Element{binary.BigEndian.Uint32(data[12:16])}
	return res
}

// zeroesImaginaryPart zeroes out the imaginary part of the Ext ensuring it
// represents a base field element.
func (z *Ext) zeroesImaginaryPart() {
	z.B0.A1.SetZero()
	z.B1.A0.SetZero()
	z.B1.A1.SetZero()
}

// BatchInvertExt inverts a slice of field extension elements
func BatchInvertExt(a []Ext) []Ext {
	casted := unsafeCast[[]Ext, []extensions.E4](&a)
	_res := extensions.BatchInvertE4(*casted)
	res := unsafeCast[[]extensions.E4, []Ext](&_res)
	return *res
}

// ParBatchInvertExt inverts a slice of field extension elements in parallel.
// Passing 0 or a negative numbers tells the function to use all the threads.
func ParBatchInvertExt(a []Ext, numCPU int) []Ext {

	if numCPU <= 0 {
		numCPU = runtime.NumCPU()
	}

	res := make([]Ext, len(a))

	parallel.Execute(len(a), func(start, stop int) {
		subRes := BatchInvertExt(a[start:stop])
		copy(res[start:stop], subRes)
	}, numCPU)

	return res
}

// SetPseudoRand sets the field element to a pseudo-random value
func (z *Ext) SetPseudoRand(rng *rand.Rand) *Ext {
	*z = PseudoRandExt(rng)
	return z
}
