// check for the logic of the design https://hackmd.io/adB-EjnKSa-P7HwI542l9g
/*warning: the implemementation does not support:
1 . the verifier should check that public-inputs are correctly related to the keccakf-state.
2. for permutation associated with the same hash, the verifer does not check the consistency of two permutations and the absorbed message
TBD : intuitive solution for later, keccakf would recive message blocks and does the XOR during aChi,
achi =2a+b+3c+2.rc would be replaced with achi = 2a+b+3c+2.(rc+msg-2.rc.msg)
the advantage is that base First,Second would not change, and third decomposition is not needed,
but the hashing proccess needs to be revised widely.
*/

package keccak

import (
	"math"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
)

// it defines keccakf permutation
func (ctx *KeccakFModule) DefineKeccakF(comp *wizard.CompiledIOP, round int) {

	// assign the names to the columns and set the parameters
	ctx.PartialKecccakFCtx()

	// pre-commit to the columns in the Ctx
	ctx.CommitKeccakFCtx(comp, round)

	//the verifier has the state in bit-base-first from the beginning
	// it checks the relation between A and ATheta; ATheta =A XOR C[i-1] XOR CC[i+1]
	ctx.addConstraintATheta(comp, round)

	//it check that each elements of  ARho is a rotation (by LR)  of corresponding element of Atheta
	ctx.addConstraintARho(comp, round)

	// it checks the relation between AChi and ARo (first comp APi from ARho);
	//AChi[i][j] =2*APi[i][j]+  APi[i+1][j] +3* APi[i+2][j]+2*RC
	ctx.addConstraintAChi(comp, round)

	// it  come back from second-base to bit-base-first (slice by slice)
	ctx.addConstraintAIota(comp, round)

	// it assignes the current LC of AIota slices to the next A
	ctx.addConstraintA(comp, round)

}

//-------------------------------------------------------------------------------------------------

// it unifies all the steps of Theta (while rotation is replaced with shift) where aThetaFirst is 65bits
func (ctx KeccakFModule) addConstraintATheta(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	var c, cc [5]*symbolic.Expression
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			c[i] = ifaces.ColumnAsVariable(h.A[i][0]).Add(ifaces.ColumnAsVariable(h.A[i][1])).Add(ifaces.ColumnAsVariable(h.A[i][2])).Add(ifaces.ColumnAsVariable(h.A[i][3])).Add(ifaces.ColumnAsVariable(h.A[i][4]))
			cc[i] = c[i].Mul(symbolic.NewConstant(First))
		}
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			eqTheta := ifaces.ColumnAsVariable(h.ATheta[i][j]).Sub(ifaces.ColumnAsVariable(h.A[i][j]).Add(c[(i-1+5)%5]).Add(cc[(i+1)%5]))
			name := ifaces.QueryIDf("Global_eqTheta_%v_%v", i, j)
			comp.InsertGlobal(round, name, eqTheta)
		}
	}
}

func (ctx KeccakFModule) addConstraintARho(comp *wizard.CompiledIOP, round int) {
	//first check decomposition of ATheta to slices
	ctx.constraintDecomposeATheta(comp, round)
	//move slice by slice, from AThetaFirst to AThetaSecond (i.e., base-first to bit-base-second) via tables Tfirst,Rsecond
	ctx.constraintAThetaFirstSecond(comp, round)
	//finally, check the relation between ARho  and AThetaSecond
	ctx.constraintAThetaSecondRho(comp, round)
}

func (ctx KeccakFModule) constraintDecomposeATheta(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	var v big.Int
	v.Exp(big.NewInt(First), big.NewInt(64), nil)

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			//aThetaFirst is 65 bits, first come back to 64 bits via committed msb
			aThetaFirst64 := ifaces.ColumnAsVariable(h.ATheta[i][j]).Sub(ifaces.ColumnAsVariable(h.msbATheta[i][j]).Mul(symbolic.NewConstant(v))).Add(ifaces.ColumnAsVariable(h.msbATheta[i][j]))
			// then compose slices
			t := ComposeHandles(h.AThetaFirstSlice[i][j], First)
			expr := aThetaFirst64.Sub(t)
			name := ifaces.QueryIDf("Global_Decomposition_ATheta_%v_%v", i, j)
			comp.InsertGlobal(round, name, expr)

		}
	}

}

// move from AThetaFirst to AThetaSecond (slice by slice)
func (ctx KeccakFModule) constraintAThetaFirstSecond(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	l := ctx.lookup
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < nS; k++ {
				name := ifaces.QueryIDf("Inclusion_CheckAThetaFirstSecond_%v_%v_%v", i, j, k)
				comp.InsertInclusion(round, name, []ifaces.Column{l.Tfirst, l.Rsecond}, []ifaces.Column{h.AThetaFirstSlice[i][j][k], h.AThetaSecondSlice[i][j][k]})

			}
		}

	}
}

// from AThetaSecond to ARo; rotation by LR
func (ctx KeccakFModule) constraintAThetaSecondRho(comp *wizard.CompiledIOP, round int) {
	//one slice is cut by the rotation, check that decomposition of  this slice is correct
	ctx.checkTargetSliceDecompose(comp, round)
	h := ctx.handle
	//finally  it rotates and converts to bit-base-second  at the same time
	ctx.checkConvertAndRotate(comp, round, h.AThetaSecondSlice, h.TargetSliceDecompos, h.ARho)
}

func (ctx KeccakFModule) checkConvertAndRotate(comp *wizard.CompiledIOP, round int, a [5][5][nS]ifaces.Column, d [5][5][nBS]ifaces.Column, res [5][5]ifaces.Column) {
	var v, u *symbolic.Expression
	var t, h, s *symbolic.Expression
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {

			if Is[i][j] == 15 && Ib[i][j] == 4 {
				t := ComposeHandles(a[i][j], Second)
				expr := ifaces.ColumnAsVariable(res[i][j]).Sub(t)
				name := ifaces.QueryIDf("Global_Decomposition_ARho_%v_%v", i, j)
				comp.InsertGlobal(round, name, expr)

			} else {
				v = symbolic.NewConstant(uint64(math.Pow(Second, float64(nBS-Ib[i][j]))))
				u = symbolic.NewConstant(uint64(math.Pow(Second, float64(nBS))))
				t = symbolic.NewConstant(0)
				if Is[i][j] != 15 {
					t = ifaces.ColumnAsVariable(a[i][j][(Is[i][j] + 1)]).Mul(v)
					for l := (Is[i][j] + 2); l < 16; l++ {
						v = v.Mul(u)
						t = t.Add(ifaces.ColumnAsVariable(a[i][j][l]).Mul(v))

					}
					v = v.Mul(u)
				}
				// v continues to grow
				for l := 0; l < Is[i][j]; l++ {
					t = t.Add(ifaces.ColumnAsVariable(a[i][j][l]).Mul(v))
					v = v.Mul(u)
				}

				//leftover of cut bits
				//if the target slice has not moved

				u = symbolic.NewConstant(Second)
				h = ifaces.ColumnAsVariable(d[i][j][0]).Mul(v)
				for l := 1; l < Ib[i][j]; l++ {
					v = v.Mul(u)
					h = h.Add(ifaces.ColumnAsVariable(d[i][j][l]).Mul(v))

				}

				// if the slice is moved
				s = symbolic.NewConstant(0)
				if Ib[i][j] != 4 {
					v = symbolic.NewConstant(1)
					s = ifaces.ColumnAsVariable(d[i][j][Ib[i][j]]).Mul(v)
					for l := Ib[i][j] + 1; l < nBS; l++ {
						v = v.Mul(u)
						s = s.Add(ifaces.ColumnAsVariable(d[i][j][l]).Mul(v))

					}
				}

				y := t.Add(h)
				x := y.Add(s)
				expr := ifaces.ColumnAsVariable(res[i][j]).Sub(x)
				name := ifaces.QueryIDf("Global_expr_ConvertRotate_%v", ifaces.ColumnAsVariable(res[i][j]))
				comp.InsertGlobal(round, name, expr)

			}
		}
	}
}

// check the target slice is correctly decomposed to bits
func (ctx KeccakFModule) checkTargetSliceDecompose(comp *wizard.CompiledIOP, round int) {
	one := symbolic.NewConstant(1)
	h := ctx.handle
	// check the target slice decomposition  (bit-base-second)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			t := symbolic.NewConstant(0)
			for k := 0; k < nBS; k++ {
				t = t.Add(ifaces.ColumnAsVariable(h.TargetSliceDecompos[i][j][k]).Mul(symbolic.NewConstant(bitPowersSecond[k])))
				// check that each chunk is a bit
				expr := ifaces.ColumnAsVariable(h.TargetSliceDecompos[i][j][k]).Mul(one.Sub(ifaces.ColumnAsVariable(h.TargetSliceDecompos[i][j][k])))
				name := ifaces.QueryIDf("Global_DecomposeIsBit_%v_%v_%v", i, j, k)
				comp.InsertGlobal(round, name, expr)
			}
			eqTargetSliceDec := ifaces.ColumnAsVariable(h.AThetaSecondSlice[i][j][Is[i][j]]).Sub(t)
			name := ifaces.QueryIDf("Global_eqTargetSliceDec_%v_%v", i, j)
			comp.InsertGlobal(round, name, eqTargetSliceDec)
		}
	}
}

// it unifies all the steps of Api,AChi and AIota
// AChi[i][j] =2*APi[i][j]+  APi[i+1][j] +3* APi[i+2][j]+2*RC (RC is from AIota, APi is obtained from ARho)
func (ctx KeccakFModule) addConstraintAChi(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	two := symbolic.NewConstant(2)
	three := symbolic.NewConstant(3)
	var eqAChiAPi *symbolic.Expression
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i == 0 && j == 0 {
				d := ifaces.ColumnAsVariable(h.RCcolumn).Mul(two)
				eqAChiAPi = ifaces.ColumnAsVariable(h.AChiSecond[j][(2*i+3*j)%5]).Sub((ifaces.ColumnAsVariable(h.ARho[i][j]).Mul(two)).Add(ifaces.ColumnAsVariable(h.ARho[(i+1)%5][(j+1)%5])).Add(ifaces.ColumnAsVariable(h.ARho[(i+2)%5][(j+2)%5]).Mul(three)).Add(d))
			} else {
				eqAChiAPi = ifaces.ColumnAsVariable(h.AChiSecond[j][(2*i+3*j)%5]).Sub((ifaces.ColumnAsVariable(h.ARho[i][j]).Mul(two)).Add(ifaces.ColumnAsVariable(h.ARho[(i+1)%5][(j+1)%5])).Add(ifaces.ColumnAsVariable(h.ARho[(i+2)%5][(j+2)%5]).Mul(three)))
			}
			name := ifaces.QueryIDf("Global_eqAChiAPi_%v_%v", i, j)
			comp.InsertGlobal(round, name, eqAChiAPi)
		}
	}
}

// it backs to bit-base-first (original form of state), slice by slice
func (ctx KeccakFModule) addConstraintAIota(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	l := ctx.lookup
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			t := ComposeHandles(h.AChiSecondSlice[i][j], Second)
			expr := ifaces.ColumnAsVariable(h.AChiSecond[i][j]).Sub(t)
			name := ifaces.QueryIDf("Global_Decomposition_AChi_%v_%v", i, j)
			comp.InsertGlobal(round, name, expr)
			for k := 0; k < nS; k++ {
				// move from base-second to bit-base-first, slice by slice
				name := ifaces.QueryIDf("Inclusion_CheckAIotaChi1_%v_%v_%v", i, j, k)
				comp.InsertInclusion(round, name, []ifaces.Column{l.Rfirst, l.Tsecond}, []ifaces.Column{h.AChiFirstSlice[i][j][k], h.AChiSecondSlice[i][j][k]})

			}
		}

	}
}

// checks that a[l]=aIota[l] (updating the state)
func (ctx KeccakFModule) addConstraintA(comp *wizard.CompiledIOP, round int) {
	h := ctx.handle
	slice := h.AChiFirstSlice
	// X is one everywhere except on period of P2nRound
	X := symbolic.NewConstant(1).Sub(variables.NewPeriodicSample(P2nRound, 0))

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			//make the linear combination, t(x) = \sum_{i,j} c_{i,j}p_{i,j}(x/w),
			//p_{i,j} the polynomial associated with slice[i][j]
			t := HandlesLC(slice[i][j], First)
			expr1 := ifaces.ColumnAsVariable(h.A[i][j]).Sub(t)
			expr := expr1.Mul(X)
			name := ifaces.QueryIDf("Global_expr_AIotaChiA_%v_%v", i, j)
			comp.InsertGlobal(round, name, expr)
		}
	}

}

// it makes a linear combination of handles
func HandlesLC(s [nS]ifaces.Column, base int) *symbolic.Expression {
	t := symbolic.NewConstant(0)
	v := symbolic.NewConstant(1)
	u := symbolic.NewConstant(uint64(math.Pow(float64(base), nBS)))

	for k := 0; k < nS; k++ {
		z := ifaces.ColumnAsVariable(column.Shift(s[k], -1)).Mul(v)
		t = t.Add(z)
		v = v.Mul(u)
	}
	return t
}

// it (De)Composes the slices in the given base
func ComposeHandles(a [nS]ifaces.Column, base int) *symbolic.Expression {
	u := symbolic.NewConstant(uint64(math.Pow(float64(base), nBS)))
	v := symbolic.NewConstant(1)
	t := symbolic.NewConstant(0)
	for k := 0; k < nS; k++ {
		t = t.Add(ifaces.ColumnAsVariable(a[k]).Mul(v))
		v = v.Mul(u)

	}
	return t
}
