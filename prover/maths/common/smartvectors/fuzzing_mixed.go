package smartvectors

import (
	"fmt"
	"math/rand/v2"

	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/linea-monorepo/prover/utils"
)

type smartVecTypeMixed int

// The order matters : combining type x with type y implies that the result
// will be of type max(x, y)
const (
	constantMixedT smartVecTypeMixed = iota
	windowMixedT
	RegularMixedT
	RotatedMixedT
)

var smartVecTypeListMixed = []smartVecTypeMixed{constantMixedT, windowMixedT, RegularMixedT, RotatedMixedT}

type testCaseMixed struct {
	name            string
	svecs           []SmartVector
	coeffs          []int
	expectedValue   SmartVector
	evaluationPoint fext.Element // Only used for polynomial evaluation
}

func (tc testCaseMixed) String() string {
	res := "Testcase:\n"
	res += "\tSVECS:\n"
	for i := range tc.svecs {
		res += fmt.Sprintf("\t\t %v : %v\n", i, tc.svecs[i].Pretty())
	}
	res += fmt.Sprintf("\tCOEFFs: %v\n", tc.coeffs)
	res += fmt.Sprintf("\tEXPECTED_VALUE: %v\n", tc.expectedValue.Pretty())
	return res
}

type testCaseGenMixed struct {
	// Randomness parameters
	seed int
	gen  *rand.Rand
	// Length and number of target vectors
	fullLen, numVec int
	// Parameters relevant for creating windows. This enforces the windows
	// to be included in a certain (which can possible roll over fullLen)
	windowWithLen        int
	windowMustStartAfter int
	// Allowed smart-vector types for this testcase
	allowedTypes []smartVecTypeMixed
}

func newTestBuilderMixed(seed int) *testCaseGenMixed {
	// Use a deterministic randomness source
	res := &testCaseGenMixed{seed: seed}
	// #nosec G404 --we don't need a cryptographic RNG for fuzzing purpose
	res.gen = rand.New(utils.NewRandSource(int64(seed)))

	// We should have some quarantee that the length is not too small
	// for the test generation
	res.fullLen = 1 << (res.gen.IntN(5) + 3)
	res.numVec = res.gen.IntN(8) + 1

	// In the test, we may restrict the inputs vectors to have a certain type
	allowedTypes := append([]smartVecTypeMixed{}, smartVecTypeListMixed...)
	res.gen.Shuffle(len(allowedTypes), func(i, j int) {
		allowedTypes[i], allowedTypes[j] = allowedTypes[j], allowedTypes[i]
	})
	res.allowedTypes = allowedTypes[:res.gen.IntN(len(allowedTypes)-1)+1]

	// Generating the window : it should be roughly half of the total length
	// this aims at maximizing the coverage.
	res.windowWithLen = res.gen.IntN(res.fullLen-4)/2 + 2
	res.windowMustStartAfter = res.gen.IntN(res.fullLen)
	return res
}
func (gen *testCaseGen) NewTestCaseForLinearCombinationMixed() (tcase testCaseMixed) {

	tcase.name = fmt.Sprintf("fuzzy-with-seed-%v-poly-eval", gen.seed)
	tcase.svecs = make([]SmartVector, gen.numVec)
	tcase.coeffs = make([]int, gen.numVec)
	tcase.evaluationPoint.SetRandom()
	x := tcase.evaluationPoint
	vals := []field.Element{}

	// MaxType is used to determine what type should the result be
	maxType := constantT

	// For the windows, we need to track the dimension of the windows
	winMinStart := gen.fullLen
	winMaxStop := 0

	for i := 0; i < gen.numVec; i++ {
		// Generate one by one the different vectors
		val := gen.genValue()
		vals = append(vals, val)
		tcase.coeffs[i] = gen.gen.IntN(10) - 5
		chosenType := gen.allowedTypes[gen.gen.IntN(len(gen.allowedTypes))]
		maxType = utils.Max(maxType, chosenType)

		switch chosenType {
		case constantT:
			tcase.svecs[i] = NewConstant(val, gen.fullLen)
		case windowT:
			v := gen.genWindow(val, val)
			tcase.svecs[i] = v
			start := normalize(v.interval().Start(), gen.windowMustStartAfter, gen.fullLen)
			winMinStart = utils.Min(winMinStart, start)

			stop := normalize(v.interval().Stop(), gen.windowMustStartAfter, gen.fullLen)
			if stop < start {
				stop += gen.fullLen
			}
			winMaxStop = utils.Max(winMaxStop, stop)
		case regularT:
			tcase.svecs[i] = gen.genRegular(val)
		case rotatedT:
			tcase.svecs[i] = gen.genRotated(val)
		}
	}

	// If there are no windows, then the initial condition that we use
	// do pass this sanity-check
	if winMaxStop-winMinStart > gen.windowWithLen {
		utils.Panic("inconsistent window dimension %v %v with gen %++v", winMinStart, winMaxStop, gen)
	}
	resVal := poly.EvalOnExtField(vals, x)

	switch {
	case maxType == constantT:
		tcase.expectedValue = NewConstantExt(resVal, gen.fullLen)
	case maxType == regularT || maxType == windowT || maxType == rotatedT:
		tcase.expectedValue = NewRegularExt(vectorext.Repeat(resVal, gen.fullLen))
	}

	return tcase
}
