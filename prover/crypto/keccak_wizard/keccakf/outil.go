package keccak

import (
	"math"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// input of hash is [5][5] elements of uint64, we convert it to [5][5] bit-base-first filed.Element
func ConvertState(a [5][5]uint64, base int) (res [5][5]field.Element) {
	z := make([]field.Element, 64)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < 64; k++ {
				z[k] = field.NewElement(a[i][j] >> k & 1)
			}
			//sanity check
			zz := Convertbase(field.NewElement(a[i][j]), 2, 64)
			if zz[i] != z[i] {
				panic("decomposition to the chunks is wrong")
			}
			if Compose(zz, 2) != field.NewElement(a[i][j]) {
				panic("state conversion is wrong")
			}
			res[i][j] = Compose(z, base)
		}
	}
	return res

}

// it compose the chunks to get the field element
func Compose(r []field.Element, base int) (res field.Element) {
	s := field.Zero()
	v := field.One()
	u := field.NewElement(uint64(base))
	var y field.Element
	for i := range r {
		y.Mul(&r[i], &v)
		s.Add(&s, &y)
		v.Mul(&u, &v)

	}
	return s
}

// it composes chunks and returns  slices
func (ctx KeccakFModule) chunkToSlice(r []field.Element, base int) (S [nS]field.Element) {
	var z field.Element
	for k := 0; k < nS; k++ {
		s := field.Zero()
		v := field.One()
		u := field.NewElement(uint64(base))

		for i := 0; i < nBS; i++ {
			z.Mul(&r[k*nBS+i], &v)
			s.Add(&s, &z)
			v.Mul(&u, &v)
		}
		S[k] = s

	}
	return S
}

// it receives a num (64 or 65 bits), and returns the representation in base "base"
func Convertbase(num field.Element, base, nb int) []field.Element {
	var len int
	r := make([]field.Element, nb)
	z := big.NewInt(int64(base))
	d := big.NewInt(0)
	m := big.NewInt(0)
	numBigInt := num.BigInt(big.NewInt(0))

	t := numBigInt
	for i := 0; i < nb-1; i++ {
		if t.Cmp(big.NewInt(int64(base))) == 1 || t.Cmp(big.NewInt(int64(base))) == 0 {
			d.DivMod(t, z, m)
			r[i].SetInterface(m)
			t = d
			len = i + 1
		}
	}

	r[len].SetInterface(t)
	if t.Cmp(big.NewInt(int64(base))) == 1 || t.Cmp(big.NewInt(int64(base))) == 0 {
		utils.Panic("decomposition in base %v is not correct, the last chunk/reminder is larger than %v", base, base)
	}
	// Sanity check
	s := Compose(r, base)
	if s != num {
		panic("The decomposition to the chunks is not correct ")
	}
	return r
}

// it receives a 65bits and outputs aReal 64 bits and its new decomposition
func backTo64(a field.Element, r []field.Element) (aReal field.Element, rNew []field.Element) {
	var z field.Element
	var v field.Element
	rnew := make([]field.Element, 64)
	v.Exp(field.NewElement(First), big.NewInt(64))

	for i := 0; i < 5; i++ {
		z.Mul(&r[64], &v)
		aReal.Sub(&a, &z).Add(&aReal, &r[64])
	}
	rnew[0].Add(&r[0], &r[64])
	for i := 1; i < 64; i++ {
		rnew[i] = r[i]
	}
	a = Compose(rnew, First)
	if a != aReal {
		panic("aReal is not correct")
	}
	return a, rnew

}

// it composes the slices to get the fieldElement
func (ctx KeccakFModule) composeSlice(a [nS]field.Element, base int) (b field.Element) {
	var z field.Element
	t := field.Zero()
	v := field.One()
	u := field.NewElement(uint64(math.Pow(float64(base), nBS)))
	for k := 0; k < nS; k++ {
		z.Mul(&a[k], &v)
		t.Add(&t, &z)
		v.Mul(&v, &u)
	}

	return t
}

// it assigns the length of witnesses
func (ctx *KeccakFModule) BuildColumns() {
	w := ctx.witness
	w.a = make([][5][5]field.Element, ctx.SIZE)

	w.aTheta = make([][5][5]field.Element, ctx.SIZE)

	w.aTheta64 = make([][5][5]field.Element, ctx.SIZE)
	w.msb = make([][5][5]field.Element, ctx.SIZE)

	w.aRho = make([][5][5]field.Element, ctx.SIZE)
	w.aPi = make([][5][5]field.Element, ctx.SIZE)
	w.aChiSecond = make([][5][5]field.Element, ctx.SIZE)
	w.aChiFirst = make([][5][5]field.Element, ctx.SIZE)

	w.aThetaFirstSlice = make([][5][5][nS]field.Element, ctx.SIZE)
	w.aThetaSecondSlice = make([][5][5][nS]field.Element, ctx.SIZE)
	w.aChiSecondSlice = make([][5][5][nS]field.Element, ctx.SIZE)
	w.aChiFirstSlice = make([][5][5][nS]field.Element, ctx.SIZE)

	w.rcFieldSecond = make([]field.Element, ctx.SIZE)
	w.aTargetSliceDecompose = make([][5][5][nBS]field.Element, ctx.SIZE)
	ctx.witness = w

}
