package multilinearmpts

import (
	"fmt"
	"math/bits"
	"sort"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/fiatshamir"
	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/sumcheck"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// entry bundles the objects the prover/verifier need for one polynomial.
type entry struct {
	view  *wiop.ColumnView  // original Lagrange column view (may carry a shift)
	point wiop.FieldPromise // eval-point coin (before shift adjustment)
	claim *wiop.Cell        // y_j = P_j(ω^s · z) cell from the LagrangeEval
}

// group collects all declared wiop objects for one MPTS compilation unit.
type group struct {
	logN       int
	n          int
	entries    []entry
	coeffMod   *wiop.Module
	coeffCols  []*wiop.Column
	lambdaCoin *wiop.CoinField
	// roundPolyCells[i] = {P(0), P(2)} for sumcheck round i.
	// The EvalGate product has degree 2; Gruen stores {P(0), P(2)}.
	roundPolyCells [][2]*wiop.Cell
	hPrimeCells    []*wiop.Cell // logN challenge cells h'[i]
	uCells         []*wiop.Cell // m individual MLE claim cells u_j = P̂_j(h')
}

// Compile performs the MPTS compilation pass on sys. For each batch of
// uncompiled LagrangeEval queries sharing the same (eval-point round, column
// size), it creates a new round with all sumcheck machinery and a
// MultilinearEval query on the coefficient columns.
//
// The sumcheck identity is Y = Σ_j λ^j Σ_h z_j^h · P̂_j[h], proved by
// folding each (P̂_j, M_j) pair (M_j[h] = z_j^h) jointly. The round
// polynomial uses the EvalGate: Σ_j λ^j · P̂_j · M_j (degree 2).
func Compile(sys *wiop.System) {
	type groupKey struct {
		roundID int
		colSize int
	}

	groups := make(map[groupKey][]entry)
	var keyOrder []groupKey

	for _, le := range sys.LagrangeEvals {
		if le.IsReduced() {
			continue
		}
		if len(le.Polynomials) == 0 {
			continue
		}
		sz := le.Polynomials[0].Column.Module.Size()
		if sz == 0 {
			panic(fmt.Sprintf(
				"mpts: module %q must be sized before calling Compile",
				le.Polynomials[0].Column.Module.Context.Path(),
			))
		}
		key := groupKey{le.Round().ID, sz}
		if _, exists := groups[key]; !exists {
			keyOrder = append(keyOrder, key)
		}
		for j, pv := range le.Polynomials {
			groups[key] = append(groups[key], entry{pv, le.EvaluationPoint, le.EvaluationClaims[j]})
		}
		le.MarkAsReduced()
	}

	sort.Slice(keyOrder, func(i, k int) bool {
		if keyOrder[i].roundID != keyOrder[k].roundID {
			return keyOrder[i].roundID < keyOrder[k].roundID
		}
		return keyOrder[i].colSize < keyOrder[k].colSize
	})

	compCtx := sys.Context.Childf("mpts")
	for gi, key := range keyOrder {
		buildGroup(sys, groups[key], key.colSize, compCtx.Childf("g%d", gi))
	}
}

func buildGroup(sys *wiop.System, entries []entry, n int, ctx *wiop.ContextFrame) {
	logN := bits.TrailingZeros(uint(n))
	m := len(entries)

	mptsRound := sys.NewRound()
	lambdaCoin := mptsRound.NewCoinField(ctx.Childf("lambda"))

	// Coefficient columns in a dedicated module.
	coeffMod := sys.NewSizedModule(ctx.Childf("coeffMod"), n, wiop.PaddingDirectionNone)
	coeffCols := make([]*wiop.Column, m)
	for j := range m {
		coeffCols[j] = coeffMod.NewColumn(
			ctx.Childf("coeff[%d]", j), wiop.VisibilityOracle, mptsRound,
		)
	}

	// Two round-poly cells per round (P(0) and P(2) in Gruen degree-2 format).
	roundPolyCells := make([][2]*wiop.Cell, logN)
	for i := range logN {
		roundPolyCells[i][0] = mptsRound.NewCell(ctx.Childf("rp[%d][0]", i), true)
		roundPolyCells[i][1] = mptsRound.NewCell(ctx.Childf("rp[%d][2]", i), true)
	}

	// One challenge cell per sumcheck round.
	hPrimeCells := make([]*wiop.Cell, logN)
	for i := range logN {
		hPrimeCells[i] = mptsRound.NewCell(ctx.Childf("h'[%d]", i), true)
	}

	// Individual MLE claim cells.
	uCells := make([]*wiop.Cell, m)
	for j := range m {
		uCells[j] = mptsRound.NewCell(ctx.Childf("u[%d]", j), true)
	}

	// MultilinearEval: coefficient columns evaluated at h'.
	evalPoints := make([]wiop.FieldPromise, logN)
	for i := range logN {
		evalPoints[i] = hPrimeCells[i]
	}
	coeffViews := make([]*wiop.ColumnView, m)
	for j := range m {
		coeffViews[j] = coeffCols[j].View()
	}
	sys.NewMultilinearEvalFrom(ctx.Childf("mle"), coeffViews, evalPoints, uCells)

	g := &group{
		logN:           logN,
		n:              n,
		entries:        entries,
		coeffMod:       coeffMod,
		coeffCols:      coeffCols,
		lambdaCoin:     lambdaCoin,
		roundPolyCells: roundPolyCells,
		hPrimeCells:    hPrimeCells,
		uCells:         uCells,
	}
	mptsRound.RegisterAction(&mptsProver{g: g})
	mptsRound.RegisterVerifierAction(&mptsVerifier{g: g})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// effectiveEvalPoint returns the evaluation point adjusted for any cyclic
// shift: P'(z) = P(ω^s · z) so the MPTS uses ω^s · z as the eval point.
func effectiveEvalPoint(e entry, n int, rt wiop.Runtime) field.Ext {
	zRaw := e.point.EvaluateSingle(rt).Value.AsExt()
	s := e.view.ShiftingOffset
	if s == 0 {
		return zRaw
	}
	omega := field.RootOfUnityBy(n)
	var omegaS field.Element
	omegaS.ExpInt64(omega, int64(s))
	var zEff field.Ext
	zEff.MulByElement(&zRaw, &omegaS)
	return zEff
}

// buildGeomMask fills dst[h] = z^h for h = 0, ..., len(dst)-1.
func buildGeomMask(dst []field.Ext, z field.Ext) {
	dst[0].SetOne()
	for h := 1; h < len(dst); h++ {
		dst[h].Mul(&dst[h-1], &z)
	}
}

// foldExt folds an ext-field table in place at challenge r: work[k] += r*(work[mid+k]-work[k]).
func foldExt(work []field.Ext, r field.Ext) {
	mid := len(work) / 2
	for k := range mid {
		var d field.Ext
		d.Sub(&work[mid+k], &work[k])
		d.Mul(&d, &r)
		work[k].Add(&work[k], &d)
	}
}

// ---------------------------------------------------------------------------
// Prover action
// ---------------------------------------------------------------------------

type mptsProver struct{ g *group }

func (p *mptsProver) Run(rt wiop.Runtime) {
	g := p.g
	n, logN, m := g.n, g.logN, len(g.entries)
	domain := fft.NewDomain(uint64(n))

	lambda := rt.GetCoinValue(g.lambdaCoin).AsExt()

	// Build λ^j powers.
	rhos := make([]field.Ext, m)
	var lambdaPow field.Ext
	lambdaPow.SetOne()
	for j := range m {
		rhos[j] = lambdaPow
		lambdaPow.Mul(&lambdaPow, &lambda)
	}

	// iNTT each Lagrange column → hat_P_j (base-field), assign coefficient col.
	hatPs := make([][]field.Element, m)
	for j, e := range g.entries {
		cv := rt.GetColumnAssignment(e.view.Column)
		mod := e.view.Column.Module
		table := make([]field.Element, n)
		for i := range n {
			table[i] = cv.ElementAtN(mod.Padding, n, i).AsBase()
		}
		domain.FFTInverse(table, fft.DIF)
		utils.BitReverse(table)
		hatPs[j] = table
		rt.AssignColumn(g.coeffCols[j], &wiop.ConcreteVector{
			Plain: field.VecFromBase(table),
		})
	}

	// Build individual geometric mask tables M_j[h] = z_j^h (ext-field).
	masks := make([][]field.Ext, m)
	for j, e := range g.entries {
		z := effectiveEvalPoint(e, n, rt)
		masks[j] = make([]field.Ext, n)
		buildGeomMask(masks[j], z)
	}

	// Lift hat_P_j to ext-field working copies (needed because the fold
	// challenge is ext-field, which promotes base × ext → ext).
	hatPsW := make([][]field.Ext, m)
	for j := range m {
		hatPsW[j] = make([]field.Ext, n)
		for h := range n {
			hatPsW[j][h] = field.Lift(hatPs[j][h])
		}
	}
	masksW := make([][]field.Ext, m)
	for j := range m {
		masksW[j] = make([]field.Ext, n)
		copy(masksW[j], masks[j])
	}

	// Compute initial claim Y = Σ_j λ^j y_j.
	var Y field.Ext
	for j, e := range g.entries {
		yj := rt.GetCellValue(e.claim).AsExt()
		var term field.Ext
		term.Mul(&rhos[j], &yj)
		Y.Add(&Y, &term)
	}

	// Virtual Fiat-Shamir.
	fs := fiatshamir.NewFiatShamir()
	fs.UpdateExt(lambda)
	fs.UpdateExt(Y)

	// EvalGate sumcheck: Y = Σ_h Σ_j λ^j hat_P_j[h] * M_j[h].
	// Round poly (degree 2): P(t) = Σ_j λ^j Σ_{k<mid} hatPj(k,t) * Mj(k,t).
	claim := Y
	hPrimeVals := make([]field.Ext, logN)

	for i := range logN {
		mid := len(hatPsW[0]) / 2

		// P(0): bottom-half sum.
		var p0 field.Ext
		for j := range m {
			for k := range mid {
				var term field.Ext
				term.Mul(&hatPsW[j][k], &masksW[j][k])
				term.Mul(&rhos[j], &term)
				p0.Add(&p0, &term)
			}
		}

		// P(2): linear-extrapolate each table to t=2, then sum.
		var p2 field.Ext
		for j := range m {
			for k := range mid {
				var h2, m2, term field.Ext
				// hatP_j at t=2: 2*top - bot
				h2.Add(&hatPsW[j][mid+k], &hatPsW[j][mid+k])
				h2.Sub(&h2, &hatPsW[j][k])
				// M_j at t=2: 2*top - bot
				m2.Add(&masksW[j][mid+k], &masksW[j][mid+k])
				m2.Sub(&m2, &masksW[j][k])

				term.Mul(&h2, &m2)
				term.Mul(&rhos[j], &term)
				p2.Add(&p2, &term)
			}
		}

		rt.AssignCell(g.roundPolyCells[i][0], field.ElemFromExt(p0))
		rt.AssignCell(g.roundPolyCells[i][1], field.ElemFromExt(p2))

		fs.UpdateExt(p0)
		fs.UpdateExt(p2)
		ri := fs.RandomFext()

		rt.AssignCell(g.hPrimeCells[i], field.ElemFromExt(ri))
		hPrimeVals[i] = ri

		// Advance claim.
		rp := sumcheck.RoundPoly{p0, p2}
		claim = rp.EvalAt(ri, claim)

		// Fold all working tables at ri.
		for j := range m {
			foldExt(hatPsW[j], ri)
			foldExt(masksW[j], ri)
			hatPsW[j] = hatPsW[j][:mid]
			masksW[j] = masksW[j][:mid]
		}
	}

	// Assign individual claims: u_j = P̂_j(h') = hatPsW[j][0].
	for j := range m {
		rt.AssignCell(g.uCells[j], field.ElemFromExt(hatPsW[j][0]))
	}
}

// ---------------------------------------------------------------------------
// Verifier action
// ---------------------------------------------------------------------------

type mptsVerifier struct{ g *group }

func (v *mptsVerifier) Check(rt wiop.Runtime) error {
	g := v.g
	logN, m := g.logN, len(g.entries)

	lambda := rt.GetCoinValue(g.lambdaCoin).AsExt()

	// Build λ^j powers and collect y_j, z_j_eff.
	rhos := make([]field.Ext, m)
	zEffs := make([]field.Ext, m)
	var Y field.Ext
	var lambdaPow field.Ext
	lambdaPow.SetOne()
	for j, e := range g.entries {
		rhos[j] = lambdaPow
		yj := rt.GetCellValue(e.claim).AsExt()
		var term field.Ext
		term.Mul(&lambdaPow, &yj)
		Y.Add(&Y, &term)
		zEffs[j] = effectiveEvalPoint(e, g.n, rt)
		lambdaPow.Mul(&lambdaPow, &lambda)
	}

	// Replay virtual FS.
	fs := fiatshamir.NewFiatShamir()
	fs.UpdateExt(lambda)
	fs.UpdateExt(Y)

	claim := Y
	hPrime := make([]field.Ext, logN)

	for i := range logN {
		p0 := rt.GetCellValue(g.roundPolyCells[i][0]).AsExt()
		p2 := rt.GetCellValue(g.roundPolyCells[i][1]).AsExt()

		fs.UpdateExt(p0)
		fs.UpdateExt(p2)
		expectedRI := fs.RandomFext()

		assignedRI := rt.GetCellValue(g.hPrimeCells[i]).AsExt()
		if !expectedRI.Equal(&assignedRI) {
			return fmt.Errorf(
				"mpts: sumcheck challenge mismatch at round %d: expected %v, got %v",
				i, expectedRI, assignedRI,
			)
		}
		hPrime[i] = expectedRI

		rp := sumcheck.RoundPoly{p0, p2}
		claim = rp.EvalAt(expectedRI, claim)
	}

	// Final check: claim == Σ_j λ^j · EvalMonomialMaskExt(z_j, h') · u_j.
	var expected field.Ext
	for j := range m {
		mj := polynomials.EvalMonomialMaskExt(zEffs[j], hPrime)
		uj := rt.GetCellValue(g.uCells[j]).AsExt()
		var term field.Ext
		term.Mul(&mj, &uj)
		term.Mul(&rhos[j], &term)
		expected.Add(&expected, &term)
	}

	if !claim.Equal(&expected) {
		return fmt.Errorf(
			"mpts: final check failed: sumcheck residual %v != Σ_j λ^j·mask_j·u_j = %v",
			claim, expected,
		)
	}
	return nil
}
