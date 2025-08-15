package field

import (
	"math"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"runtime"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

type Vector []Fr
type Octuplet [8]Fr

const (
	// RmaxOrderRoot
	MaxOrderRoot uint64 = 24
	// MultiplicativeGen generator of 𝔽ᵣ*
	MultiplicativeGen uint64 = 3
	// number of 32 bits words needed to represent a Fr
	Limbs = 1
	// Bits is the number of bits needed to represent a field element.
	Bits = koalabear.Bits
	// Bytes is the number of bytes needed to represent a field element.
	Bytes = koalabear.Bytes
)

var (
	RootOfUnity     = StringToFr("1791270792")
	Modulus         = koalabear.Modulus
	Butterfly       = koalabear.Butterfly
	montConstant    = StringToFr("33554430")
	montConstantInv = StringToFr("1057030144")
)

// NewFr returns a new field element from a uint64
func NewFr(v uint64) Fr {
	res := koalabear.NewElement(v)
	return Fr(res)
}

// Zero returns the zero field element
func Zero[T anyF]() T {
	var res T
	return res
}

// One returns the one field element
func One[T anyF, PT interface {
	SetOne() *T
	*T
}]() T {
	var res T
	PT(&res).SetOne()
	return res
}

// MulByMontgommeryConst multiplies by montConstant, where montConstant is the Montgommery constant
func (z *Fr) MulByMontgommeryConst(x *Fr) *Fr {
	z.Mul(x, &montConstant)
	return z
}

// DivByMontgommeryConst multiplies the field element by R^-1, where R is the Montgommery constant
func (z *Fr) DivByMontgommeryConst(x *Fr) *Fr {
	z.Mul(x, &montConstantInv)
	return z
}

// StringToFr constructs a new field element from a string. The rules to
// determine how the string is casted into a field elements are the one of
// [Fr.SetString]
func StringToFr(s string) (res Fr) {
	_, err := res.SetString(s)
	if err != nil {
		panic(err)
	}
	return res
}

// ToInt converts a field element to an int
func (e *Fr) ToInt() int {
	n := e.Uint64()
	if !e.IsUint64() || n > math.MaxInt {
		panic("out of range")
	}
	return int(n) // #nosec G115 -- Checked for overflow
}

// ExpToInt sets z to x**k
func (z *Fr) ExpToInt(x Fr, k int) *Fr {
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

// RandomFr returns a random element
func RandomFr() Fr {
	var res Fr
	res.SetRandom()
	return res
}

// PseudoRand generates a field using a pseudo-random number generator
func PseudoRand(rng *rand.Rand) Fr {

	var (
		bigInt    = &big.Int{}
		res       = Fr{}
		bareU32   = [1]uint32{rng.Uint32()}
		bareBytes = *(*[4]byte)(unsafe.Pointer(&bareU32))
	)

	bigInt.SetBytes(bareBytes[:]).Mod(bigInt, Modulus())
	res.SetBigInt(bigInt)
	return res
}

// SetPseudoRand sets the field element to a pseudo-random value
func (z *Fr) SetPseudoRand(rng *rand.Rand) *Fr {
	*z = PseudoRand(rng)
	return z
}

// PseudoRandTruncated generates a field using a pseudo-random number generator
func PseudoRandTruncated(rng *rand.Rand, sizeByte int) Fr {

	if sizeByte > 4 {
		utils.Panic("supplied a byteSize larger than 4 (%v), this must be a mistake. Please check that the supplied value is not instead a BIT-size.", sizeByte)
	}

	var (
		bigInt    = &big.Int{}
		res       = Fr{}
		bareU32   = [1]uint32{rng.Uint32()}
		bareBytes = *(*[4]byte)(unsafe.Pointer(&bareU32))
	)

	bigInt.SetBytes(bareBytes[:sizeByte]).Mod(bigInt, Modulus())
	res.SetBigInt(bigInt)
	return res
}

// BoolTo returns 1 if true and zero if false
func BoolTo[F anyF, PT interface {
	SetOne() *F
	*F
}](b bool) F {
	if b {
		return One[F, PT]()
	}
	return Zero[F]()
}

// ParseOctuplet parses a [32]byte into an octuplet of field element
func ParseOctuplet(data [32]byte) [8]Fr {
	var res [8]Fr
	for i := range res {
		if err := res[i].SetBytesCanonical(data[i*4 : i*4+4]); err != nil {
			utils.Panic("could not parse octuplet: %v, data: %v", err, data)
		}
	}
	return res
}

// OctupletToBytes converts an octuplet of field elements to a [32]byte
func OctupletToBytes(octuplet [8]Fr) [32]byte {
	var res [32]byte
	for i := range octuplet {
		x := octuplet[i].Bytes()
		copy(res[i*4:i*4+4], x[:])
	}
	return res
}

// BatchInvertFr inverts a slice of field elements
func BatchInvertFr(a []Fr) []Fr {
	casted := unsafeCast[[]Fr, []koalabear.Element](&a)
	_res := koalabear.BatchInvert(*casted)
	res := unsafeCast[[]koalabear.Element, []Fr](&_res)
	return *res
}

// ParBatchInvert is as a parallel implementation of [BatchInvert]. The caller
// can supply the target number of cores to use to perform the paralellization.
// If `numCPU=0` is provided, the function defaults to using all the available
// cores exposed by the OS.
func ParBatchInvertFr(a []Fr, numCPU int) []Fr {

	if numCPU == 0 {
		numCPU = runtime.NumCPU()
	}

	res := make([]Fr, len(a))

	parallel.Execute(len(a), func(start, stop int) {
		subRes := BatchInvertFr(a[start:stop])
		copy(res[start:stop], subRes)
	}, numCPU)

	return res
}

// AddFr is as Add and is implemented for genericity reasons
func (z *Fr) AddFr(x, y Fr) *Fr {
	return z.Add(&x, &y)
}
