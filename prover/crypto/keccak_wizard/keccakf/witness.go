package keccak

import (
	"fmt"
	"math"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// it receives 'a' and return aTetaFirst (all the steps of Theta are unified here)
func (ctx KeccakFModule) assignAThetaFirst(a [5][5]field.Element) (aa [5][5]field.Element) {
	var c, cc [5]field.Element
	firstField := field.NewElement(First)
	for i := 0; i < 5; i++ {
		c[i].Add(&a[i][0], &a[i][1]).Add(&c[i], &a[i][2]).Add(&c[i], &a[i][3]).Add(&c[i], &a[i][4])
		cc[i].Mul(&firstField, &c[i])
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			// aa is 65 bits now
			aa[i][j].Add(&a[i][j], &c[(i-1+5)%5]).Add(&aa[i][j], &cc[(i+1)%5])

		}
	}

	return aa
}

// it first come back to Length=64 bits and then finds the slices of aTheta
func (ctx KeccakFModule) assignAThetaSlice(a [5][5]field.Element) (aTheta [5][5]field.Element, sliceFirst, sliceSecond [5][5][nS]field.Element, targetSliceDecompose [5][5][nBS]field.Element, targetSlice [5][5]field.Element) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {

			aTheta[i][j], sliceFirst[i][j], sliceSecond[i][j], targetSliceDecompose[i][j], targetSlice[i][j] = ctx.fieldSliceFirstToSecond(a[i][j], Is[i][j])

		}

	}
	return aTheta, sliceFirst, sliceSecond, targetSliceDecompose, targetSlice
}

// it rotate each chunk in bit-base-second by LR position (via Is,Ib)
func (ctx KeccakFModule) assignARho(a [5][5][nS]field.Element, d [5][5][nBS]field.Element) (res [5][5]field.Element) {
	var z, v, u field.Element
	var t, h, s field.Element
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {

			if Is[i][j] == 15 && Ib[i][j] == 4 {
				res[i][j] = ctx.composeSlice(a[i][j], Second)
			} else {
				v = field.NewElement(uint64(math.Pow(Second, float64(nBS-Ib[i][j]))))
				u = field.NewElement(uint64(math.Pow(Second, float64(nBS))))
				t = field.Zero()
				if Is[i][j] != 15 {
					t.Mul(&a[i][j][(Is[i][j]+1)], &v)
					for l := (Is[i][j] + 2); l < 16; l++ {
						v.Mul(&v, &u)
						z.Mul(&a[i][j][l], &v)
						t.Add(&t, &z)

					}
					v.Mul(&v, &u)
				}
				// v continues to grow
				for l := 0; l < Is[i][j]; l++ {
					z.Mul(&a[i][j][l], &v)
					t.Add(&t, &z)
					v.Mul(&v, &u)

				}

				//leftover of cut bits

				u = field.NewElement(Second)
				h.Mul(&d[i][j][0], &v)
				for l := 1; l < Ib[i][j]; l++ {
					v.Mul(&v, &u)
					z.Mul(&d[i][j][l], &v)
					h.Add(&h, &z)

				}
				s = field.Zero()
				if Ib[i][j] != 4 {
					v = field.One()
					s.Mul(&d[i][j][Ib[i][j]], &v)
					for l := Ib[i][j] + 1; l < nBS; l++ {
						v.Mul(&v, &u)
						z.Mul(&d[i][j][l], &v)
						s.Add(&s, &z)

					}
				}

				z.Add(&t, &h)
				res[i][j].Add(&z, &s)
			}

		}

	}
	return res
}

// lane shuffles over ARho
func (ctx KeccakFModule) assignAPi(aRho [5][5]field.Element) (aPi [5][5]field.Element) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			aPi[j][(2*i+3*j)%5] = aRho[i][j]

		}
	}
	return aPi
}

// AChiSecond[i][j] =2*APi[i][j]+  APi[i+1][j] +3* APi[(i+2)% 5][j]+2*RC
func (ctx KeccakFModule) assignAChiArith(aPi [5][5]field.Element, rc field.Element) (aChiArith [5][5]field.Element) {
	var z, y, u field.Element
	two := field.NewElement(2)
	three := field.NewElement(3)
	var t [5][5]field.Element
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			z.Mul(&aPi[i][j], &two)
			u.Add(&z, &aPi[(i+1)%5][j])
			y.Mul(&aPi[(i+2)%5][j], &three)
			t[i][j].Add(&u, &y)
			if i == 0 && j == 0 {
				z.Mul(&rc, &two)

				t[i][j].Add(&t[i][j], &z)
			}

		}
	}
	return t
}

// takes aChi in base second, and gives the slices in base second,first, and aChi in base first
func (ctx KeccakFModule) assignAIotaChi(aChiSecond [5][5]field.Element) (sliceSecond, sliceFirst [5][5][nS]field.Element, aChiFirst [5][5]field.Element) {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {

			sliceSecond[i][j], sliceFirst[i][j], aChiFirst[i][j] = ctx.fieldSliceSecondToFirst(aChiSecond[i][j])

		}
	}
	return sliceSecond, sliceFirst, aChiFirst
}

// it assign the witness for RCcolumn
func (ctx KeccakFModule) assignRCfieldSecond(RC [nRound]uint64, l int) (h field.Element) {

	// bit decomposition of RC
	r := make([]field.Element, Length)
	if l < nRound {
		z := field.Zero()
		v := field.One()
		s := field.Zero()
		u := field.NewElement(Second)
		for k := 0; k < Length; k++ {
			r[k] = field.NewElement(RC[l] >> k & 1)
			z.Mul(&r[k], &v)
			s.Add(&s, &z)
			v.Mul(&v, &u)
		}
		h = s
	} else {
		h = field.Zero()
	}
	return h
}

// commitment to the witness-columns

func (ctx KeccakFModule) assignCommitColumnState(run *wizard.ProverRuntime, a [][5][5]field.Element, name [5][5]ifaces.ColID) {
	var aa [5][5][]field.Element
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			aa[i][j] = make([]field.Element, ctx.SIZE)
			for l := 0; l < ctx.SIZE; l++ {
				aa[i][j][l] = a[l][i][j]
			}
			run.AssignColumn(name[i][j], smartvectors.NewRegular(aa[i][j][:]))

		}
	}

}

func (ctx KeccakFModule) assignCommitColumnSlice(run *wizard.ProverRuntime, a [][5][5][nS]field.Element, name [5][5][nS]ifaces.ColID) {
	var aa [5][5][nS][]field.Element
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nS; k++ {
				aa[i][j][k] = make([]field.Element, ctx.SIZE)
				for l := 0; l < ctx.SIZE; l++ {
					aa[i][j][k][l] = a[l][i][j][k]
				}
				run.AssignColumn(name[i][j][k], smartvectors.NewRegular(aa[i][j][k][:]))

			}
		}
	}

}
func (ctx KeccakFModule) assignCommitColumnDecompose(run *wizard.ProverRuntime, a [][5][5][nBS]field.Element, name [5][5][nBS]ifaces.ColID) {
	var aa [5][5][nBS][]field.Element
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nBS; k++ {
				aa[i][j][k] = make([]field.Element, ctx.SIZE)
				for l := 0; l < ctx.SIZE; l++ {
					aa[i][j][k][l] = a[l][i][j][k]
				}
				run.AssignColumn(name[i][j][k], smartvectors.NewRegular(aa[i][j][k][:]))

			}
		}
	}

}

func (ctx KeccakFModule) assignCommitColumnRC(run *wizard.ProverRuntime, rc []field.Element, name ifaces.ColID) {

	run.AssignColumn(name, smartvectors.NewRegular(rc[:]))

}

func (ctx KeccakFModule) BuildTableFirstToSecond(run *wizard.ProverRuntime) {
	cn := ctx.colName
	t := make([]field.Element, tableSizeFirstToSecond)
	h := make([]field.Element, tableSizeFirstToSecond)
	if nBS != 4 {
		panic(fmt.Sprintf("Table Tbinary needs nBS=4, but nBS is %v; change the  table", nBS))
	}
	realSize := int(math.Pow(First, 4))
	for l0 := 0; l0 < First; l0++ {
		for l1 := 0; l1 < First; l1++ {
			for l2 := 0; l2 < First; l2++ {
				for l3 := 0; l3 < First; l3++ {
					r := l0 + l1*First + l2*int(math.Pow(First, 2)) + l3*int(math.Pow(First, 3))
					t[r] = field.NewElement(uint64(r))
					rr := convertIntToBit(l0) + convertIntToBit(l1)*Second + convertIntToBit(l2)*int(math.Pow(Second, 2)) + convertIntToBit(l3)*int(math.Pow(Second, 3))
					h[r] = field.NewElement(uint64(rr))
				}
			}
		}
	}

	for l := realSize; l < tableSizeFirstToSecond; l++ {
		t[l] = t[0]
		h[l] = h[0]
	}

	run.AssignColumn(cn.NameTfirst, smartvectors.NewRegular(t[:]))
	run.AssignColumn(cn.NameRsecond, smartvectors.NewRegular(h[:]))
}

func (ctx KeccakFModule) BuildTableSecondToFirst(run *wizard.ProverRuntime) {
	cn := ctx.colName
	t := make([]field.Element, tableSizeSecondToFirst)
	h := make([]field.Element, tableSizeSecondToFirst)
	if nBS != 4 {
		panic(fmt.Sprintf("Table Tbinary needs nBS=4, but nBS is %v; change the  table", nBS))
	}
	realSize := int(math.Pow(Second, 4))
	for l0 := 0; l0 < Second; l0++ {
		for l1 := 0; l1 < Second; l1++ {
			for l2 := 0; l2 < Second; l2++ {
				for l3 := 0; l3 < Second; l3++ {
					r := l0 + l1*Second + l2*int(math.Pow(Second, 2)) + l3*int(math.Pow(Second, 3))
					t[r] = field.NewElement(uint64(r))
					rr := convertBinaryToBit(l0) + convertBinaryToBit(l1)*First + convertBinaryToBit(l2)*int(math.Pow(First, 2)) + convertBinaryToBit(l3)*int(math.Pow(First, 3))
					h[r] = field.NewElement(uint64(rr))
				}
			}
		}
	}

	for l := realSize; l < tableSizeSecondToFirst; l++ {
		t[l] = t[0]
		h[l] = h[0]
	}

	run.AssignColumn(cn.NameTsecond, smartvectors.NewRegular(t[:]))
	run.AssignColumn(cn.NameRfirst, smartvectors.NewRegular(h[:]))
}
func convertBinaryToBit(a int) (res int) {
	if a == 0 || a == 1 || a == 4 || a == 5 || a == 8 {
		res = 0
	}
	if a == 2 || a == 3 || a == 6 || a == 7 {
		res = 1
	}
	return res
}
func convertToBit(r field.Element) (res field.Element) {
	for i := 0; i < First; i++ {
		if r == field.NewElement(uint64(i)) {
			if i%2 == 0 {
				res = field.Zero()
			} else {
				res = field.One()
			}
		}
	}
	return res
}
func convertIntToBit(r int) (res int) {
	if r%2 == 0 {
		res = 0
	} else {
		res = 1
	}
	return res

}

func convertToBitByTable(b field.Element) (bb field.Element) {
	Two := field.NewElement(2)
	Three := field.NewElement(3)
	Four := field.NewElement(4)
	Five := field.NewElement(5)
	Six := field.NewElement(6)
	Seven := field.NewElement(7)
	Eigth := field.NewElement(8)
	switch b {
	case field.Zero():
		bb = field.Zero()
	case field.One():
		bb = field.Zero()
	case Two:
		bb = field.One()
	case Three:
		bb = field.One()
	case Four:
		bb = field.Zero()
	case Five:
		bb = field.Zero()
	case Six:
		bb = field.One()
	case Seven:
		bb = field.One()
	case Eigth:
		bb = field.Zero()

	default:
		panic(fmt.Sprintf("it is not in the correct range  input is larger than  base second =%v ", Second))
	}

	return bb
}

// 'a' is Length+1=65 bits, msb is the msb of 'a', but aTheta,S,B,D corespond with 64bit version
func (ctx KeccakFModule) fieldSliceFirstToSecond(a field.Element, Is int) (aTheta field.Element, S [nS]field.Element, B [nS]field.Element, D [nBS]field.Element, msb field.Element) {
	rr := make([]field.Element, Length)

	//get the chuncks for 65 bit
	r := Convertbase(a, First, Length+1)
	aTheta, rNew := backTo64(a, r)
	msb = r[Length]
	for i := range rNew {
		rr[i] = convertToBit(rNew[i])
	}
	//decomposition of target slice in bit-base-second
	for j := 0; j < nBS; j++ {
		D[j] = rr[Is*nBS+j]
	}
	sliceFirst := ctx.chunkToSlice(rNew, First)
	sliceSecond := ctx.chunkToSlice(rr, Second)
	// sanity check
	t := ctx.composeSlice(sliceFirst, First)
	if t != aTheta {
		panic(fmt.Sprintf("The decomposition to the  slices is not correct aReal=%v and t=%v", aTheta, t))
	}
	if t != Compose(rNew, First) {
		panic("composeSlice is corrupted")
	}

	return aTheta, sliceFirst, sliceSecond, D, msb
}

// it returns slices in second-base, bit-base-first and also aChi in bit-base-first
func (ctx KeccakFModule) fieldSliceSecondToFirst(a field.Element) (sliceSecond, sliceFirst [nS]field.Element, aChiFirst field.Element) {
	rr := make([]field.Element, Length)
	// decomposition in base-second
	r := Convertbase(a, Second, Length)

	for i := range r {
		// it convert the decomposition to bits, preparing itself for bit-base-first
		rr[i] = convertToBitByTable(r[i])
	}
	if len(r) == Length {
		// by the table Tbinary zero is equivalent with zero, so all good for zero bits between len(r) and Length=64
		//build the slices of aChiArith in base-second
		sliceSecond = ctx.chunkToSlice(r, Second)
		// move to bit-base-first slice by slice
		sliceFirst = ctx.chunkToSlice(rr, First)
	} else {
		utils.Panic("input should have %v chuncks", Length)
	}
	// sanity check
	t := ctx.composeSlice(sliceSecond, Second)
	if t != a {
		panic("The decomposition to the  slices is not correct")
	}
	//compose the slices to get AIota
	aChiFirst = ctx.composeSlice(sliceFirst, First)

	return sliceSecond, sliceFirst, aChiFirst

}
