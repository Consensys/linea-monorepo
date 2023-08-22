package keccak

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

const (
	// First base, second base
	First  = 12
	Second = 9
	//number of slices for base-first and base-second
	nS = 16
	//number of chunks per slice (chunks are the coefficients in the chosen base)
	nBS = 64 / nS
	// number of rounds in keccakf
	nRound   = 24
	P2nRound = 32 // next power of two for nRound
	Length   = 64 // bit length of original-input (uint64) in keccakf
)

// size of the lookup tables
const (
	// real size is First**nBS, but we need the next power of two
	tableSizeFirstToSecond = 1 << 15
	//real size is Second**nBS
	tableSizeSecondToFirst = 1 << 13
)

// left Rotations in Mod 64
var LR = [5][5]int{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

// slice-index cut by left rotation
var Is = [5][5]int{
	{15, 6, 15, 5, 11},
	{15, 4, 13, 4, 15},
	{0, 14, 5, 12, 0},
	{8, 2, 9, 10, 1},
	{9, 10, 6, 13, 12},
}

// bit-index cut by left rotation
var Ib = [5][5]int{
	{4, 4, 1, 3, 2},
	{3, 4, 2, 3, 2},
	{2, 2, 1, 1, 3},
	{4, 1, 3, 3, 4},
	{1, 4, 1, 4, 2},
}

var bitPowersSecond = [nBS]int{
	1, 9, 81, 729,
}

// round constant of keccakf
var RC = [24]uint64{
	0x0000000000000001,
	0x0000000000008082,
	0x800000000000808A,
	0x8000000080008000,
	0x000000000000808B,
	0x0000000080000001,
	0x8000000080008081,
	0x8000000000008009,
	0x000000000000008A,
	0x0000000000000088,
	0x0000000080008009,
	0x000000008000000A,
	0x000000008000808B,
	0x800000000000008B,
	0x8000000000008089,
	0x8000000000008003,
	0x8000000000008002,
	0x8000000000000080,
	0x000000000000800A,
	0x800000008000000A,
	0x8000000080008081,
	0x8000000000008080,
	0x0000000080000001,
	0x8000000080008008,
}

// The context for  keccakF
type KeccakFModule struct {
	SIZE int //size of the columns
	NP   int //number of permutations in a batch

	handle  Handle
	colName Name
	witness Witness
	lookup  LookUpTable

	//publicInput has two parts;  InpitPI := input of permutations, and OutputPI := output of permutations
	InputPI, OutputPI [][5][5]field.Element
}

// lookuptables, can be preprocessed
type LookUpTable struct {
	//for bit-base tables use R and for standard base table use T
	/* Tables:
	Tfirst base-first
	Rfirst bit-base-first
	Rsecond bit-base-second
	Tsecond base-second
	*/
	Tfirst, Rfirst, Rsecond, Tsecond ifaces.Column
}

// list of columns associated with the witness
type Handle struct {
	// the state in the form bit-base-first
	A [5][5]ifaces.Column
	//state after theta step
	ATheta    [5][5]ifaces.Column
	msbATheta [5][5]ifaces.Column
	//slices of ATheta in base-first
	AThetaFirstSlice [5][5][nS]ifaces.Column
	//slices of ATheta in bit-base-second
	AThetaSecondSlice [5][5][nS]ifaces.Column
	//the  decomposition of slice cut by the rotation LR
	TargetSliceDecompos [5][5][nBS]ifaces.Column

	//  state converted to bit-base-second and rotated by LR
	ARho [5][5]ifaces.Column

	//Complex binary replaced by arithmetics 2a+b+3c+2d
	AChiSecond [5][5]ifaces.Column
	// slices in base-second
	AChiSecondSlice [5][5][nS]ifaces.Column
	// back to bit-base-first (the original form of the state)
	AChiFirstSlice [5][5][nS]ifaces.Column

	// column of round constants
	RCcolumn ifaces.Column
}

// names for the columns
type Name struct {
	NameA, NameAThetaFirst, NameARho, NamemsbAThetaFirst                               [5][5]ifaces.ColID
	NameAChiArith, NameAIotaChi                                                        [5][5]ifaces.ColID
	NameAThetaFirstSlice, NameAThetaSecondSlice, NameAChiArithSlice, NameAIotaChiSlice [5][5][nS]ifaces.ColID
	NameTargetSliceDecompos                                                            [5][5][nBS]ifaces.ColID
	NameTfirst, NameRsecond, NameTsecond                                               ifaces.ColID
	NameRfirst, NameRCcolumn, NameOneColumn                                            ifaces.ColID
}

// witnesses for the columns
type Witness struct {
	a, aTheta, aTheta64, msb            [][5][5]field.Element
	aRho, aPi, aChiSecond               [][5][5]field.Element
	aChiFirst                           [][5][5]field.Element
	aThetaFirstSlice, aThetaSecondSlice [][5][5][nS]field.Element
	aChiSecondSlice, aChiFirstSlice     [][5][5][nS]field.Element
	rcFieldSecond                       []field.Element
	aTargetSliceDecompose               [][5][5][nBS]field.Element
}

// Naming the columns of Module,NP should already be set
func (ctx *KeccakFModule) PartialKecccakFCtx() {

	var nameA, nameAThetaFirst, nameARho, namemsbAThetaFirst [5][5]ifaces.ColID
	var nameAChiArith, nameAIotaChi, nameAThetaTargetSlice [5][5]ifaces.ColID
	var nameAThetaFirstSlice, nameAThetaSecondSlice, nameAChiArithSlice, nameAIotaChiSlice [5][5][nS]ifaces.ColID
	var nameTargetSliceDecompos [5][5][nBS]ifaces.ColID

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			nameA[i][j] = deriveName("A", i, j, 0)
			nameAThetaFirst[i][j] = deriveName("AThetaFirst", i, j, 0)
			nameARho[i][j] = deriveName("ARho", i, j, 0)
			nameAChiArith[i][j] = deriveName("AChiArith", i, j, 0)
			nameAThetaTargetSlice[i][j] = deriveName("AThetaTargetSlice", i, j, 0)
			namemsbAThetaFirst[i][j] = deriveName("msbAThetaFirst", i, j, 0)

		}
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nS; k++ {
				nameAThetaFirstSlice[i][j][k] = deriveName("AThetaFirstSlice", i, j, k)
				nameAThetaSecondSlice[i][j][k] = deriveName("AThetaSecondSlice", i, j, k)
				nameAChiArithSlice[i][j][k] = deriveName("AChiArithSlice", i, j, k)
				nameAIotaChiSlice[i][j][k] = deriveName("AIotaChiSlice", i, j, k)

			}

		}
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nBS; k++ {
				nameTargetSliceDecompos[i][j][k] = deriveName("TargetSliceDecompos", i, j, k)
			}
		}
	}

	size := ctx.NP * P2nRound
	colName := Name{
		NameA:                   nameA,
		NameAThetaFirst:         nameAThetaFirst,
		NameARho:                nameARho,
		NameAChiArith:           nameAChiArith,
		NameAIotaChi:            nameAIotaChi,
		NameAThetaFirstSlice:    nameAThetaFirstSlice,
		NameAThetaSecondSlice:   nameAThetaSecondSlice,
		NameAChiArithSlice:      nameAChiArithSlice,
		NameAIotaChiSlice:       nameAIotaChiSlice,
		NameTargetSliceDecompos: nameTargetSliceDecompos,
		NameRCcolumn:            deriveName("RCcolumn", 0, 0, 0),
		NameTfirst:              deriveName("Tfirst", 0, 0, 0),
		NameRsecond:             deriveName("Rsecond", 0, 0, 0),
		NameTsecond:             deriveName("Tsecond", 0, 0, 0),
		NameRfirst:              deriveName("Rfirst", 0, 0, 0),
		NamemsbAThetaFirst:      namemsbAThetaFirst,
	}
	ctx.InputPI = make([][5][5]field.Element, ctx.NP)
	ctx.OutputPI = make([][5][5]field.Element, ctx.NP)

	ctx.SIZE = size
	ctx.colName = colName

}

// pre-commit to the columns of Module
func (ctx *KeccakFModule) CommitKeccakFCtx(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	cn := ctx.colName
	l := ctx.lookup
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			h.A[i][j] = comp.InsertCommit(round, cn.NameA[i][j], ctx.SIZE)
			h.ATheta[i][j] = comp.InsertCommit(round, cn.NameAThetaFirst[i][j], ctx.SIZE)
			h.msbATheta[i][j] = comp.InsertCommit(round, cn.NamemsbAThetaFirst[i][j], ctx.SIZE)
			h.ARho[i][j] = comp.InsertCommit(round, cn.NameARho[i][j], ctx.SIZE)
			h.AChiSecond[i][j] = comp.InsertCommit(round, cn.NameAChiArith[i][j], ctx.SIZE)

		}
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nS; k++ {
				h.AThetaFirstSlice[i][j][k] = comp.InsertCommit(round, cn.NameAThetaFirstSlice[i][j][k], ctx.SIZE)
				h.AThetaSecondSlice[i][j][k] = comp.InsertCommit(round, cn.NameAThetaSecondSlice[i][j][k], ctx.SIZE)
				h.AChiSecondSlice[i][j][k] = comp.InsertCommit(round, cn.NameAChiArithSlice[i][j][k], ctx.SIZE)
				h.AChiFirstSlice[i][j][k] = comp.InsertCommit(round, cn.NameAIotaChiSlice[i][j][k], ctx.SIZE)

			}

		}
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nBS; k++ {

				h.TargetSliceDecompos[i][j][k] = comp.InsertCommit(round, cn.NameTargetSliceDecompos[i][j][k], ctx.SIZE)
			}
		}
	}

	//pre-commitment to the tables
	l.Tfirst = comp.InsertCommit(round, cn.NameTfirst, tableSizeFirstToSecond)
	l.Rsecond = comp.InsertCommit(round, cn.NameRsecond, tableSizeFirstToSecond)
	l.Tsecond = comp.InsertCommit(round, cn.NameTsecond, tableSizeSecondToFirst)
	l.Rfirst = comp.InsertCommit(round, cn.NameRfirst, tableSizeSecondToFirst)
	h.RCcolumn = comp.InsertCommit(round, cn.NameRCcolumn, ctx.SIZE)
	ctx.handle = h
	ctx.lookup = l
}
func deriveName(keccakCtx string, i, j, k int) ifaces.ColID {
	return ifaces.ColIDf("%v_%v_%v_%v_%v", "KeccakF", keccakCtx, i, j, k)
}
