package distributed

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// ConglomeratorCompilation hold the compilation context of the conglomeration
// proof. It stores pointers to the type of proof it can conglomerate and
// pointers of the resulting compiled IOP object.
type ConglomeratorCompilation struct {

	// MaxNbProofs is the maximum number of proofs that can be conglomerated
	// by the conglomeration proof at once.
	MaxNbProofs int

	// ModuleProofs lists the wizard whose proof are supported by the current
	// instance of the conglomerator.
	ModuleProofs []*wizard.CompiledIOP

	// Wiop is the compiled IOP of the conglomeration wizard.
	Wiop *wizard.CompiledIOP

	// Recursion is the recursion context used to compile the conglomeration
	// proof.
	Recursion *recursion.Recursion
}

// ConglomerationHolisticCheck is a [wizard.VerifierAction] checking that all
// the public inputs of the subproofs are the right ones.
type ConglomerateHolisticCheck struct {
	ConglomeratorCompilation
}

// Conglomerate constructs and returns a new ConglomeratorCompilation object.
// The Wiop of the returned object is compiled with iterative layers of
// self-recursion.
func Conglomerate(maxNbProofs int, moduleProofs []*wizard.CompiledIOP) *ConglomeratorCompilation {

	return &ConglomeratorCompilation{
		MaxNbProofs:  maxNbProofs,
		ModuleProofs: moduleProofs,
		Wiop:         nil,
	}

}

// Compile compiles the conglomeration proof. The function first checks if the public
// inputs are compatible and then compiles the conglomeration proof.
func (c *ConglomeratorCompilation) Compile(comp *wizard.CompiledIOP) {

	w0 := c.ModuleProofs[0]

	for i := 1; i < len(c.ModuleProofs); i++ {

		diff1, diff2 := cmpWizardIOP(w0, c.ModuleProofs[i])
		if len(diff1) > 0 || len(diff2) > 0 {
			utils.Panic("incompatible IOPs +++=%v\n---=%v", diff1, diff2)
		}
	}

	c.Recursion = recursion.DefineRecursionOf(comp, w0, recursion.Parameters{
		Name:                   "conglomeration",
		WithoutGkr:             true,
		MaxNumProof:            c.MaxNbProofs,
		WithExternalHasherOpts: true,
	})

}

// Run implements the [wizard.VerifierAction] interface.
func (c *ConglomerateHolisticCheck) Run(run_ ifaces.Runtime) error {

	var (
		run                 = run_.(wizard.Runtime)
		allGrandProduct     = field.NewElement(1)
		allLogDerivativeSum = field.Element{}
		allHornerSum        = field.Element{}
		prevGlobalSent      = field.Element{}
		prevHornerN1Hash    = field.Element{}
		merkleRootsLPP      = []field.Element{}
		merkleRootsGL       = []field.Element{}
		mainErr             error
	)

	for i := 0; i < c.MaxNbProofs; i++ {

		var (
			verifyingKey           = c.Recursion.GetPublicInputOfInstance(run, verifyingKeyPublicInput, i)
			logDerivativeSum       = c.Recursion.GetPublicInputOfInstance(run, logDerivativeSumPublicInput, i)
			grandProduct           = c.Recursion.GetPublicInputOfInstance(run, grandProductPublicInput, i)
			hornerSum              = c.Recursion.GetPublicInputOfInstance(run, hornerPublicInput, i)
			hornerN0Hash           = c.Recursion.GetPublicInputOfInstance(run, hornerN0HashPublicInput, i)
			hornerN1Hash           = c.Recursion.GetPublicInputOfInstance(run, hornerN1HashPublicInput, i)
			globalReceived         = c.Recursion.GetPublicInputOfInstance(run, globalReceiverPublicInput, i)
			globalSent             = c.Recursion.GetPublicInputOfInstance(run, globalSenderPublicInput, i)
			isFirst                = c.Recursion.GetPublicInputOfInstance(run, isFirstPublicInput, i)
			isLast                 = c.Recursion.GetPublicInputOfInstance(run, isLastPublicInput, i)
			isLPP                  = c.Recursion.GetPublicInputOfInstance(run, isLppPublicInput, i)
			nbActualLppPublicInput = c.Recursion.GetPublicInputOfInstance(run, nbActualLppPublicInput, i)

			prevVerifyingKey, nextVerifyingKey             field.Element
			sameVerifyingKeyAsPrev, sameVerifyingKeyAsNext bool
		)

		if i > 0 {
			prevVerifyingKey = c.Recursion.GetPublicInputOfInstance(run, verifyingKeyPublicInput, i-1)
			sameVerifyingKeyAsPrev = verifyingKey == prevVerifyingKey
		}

		if i < c.MaxNbProofs-1 {
			nextVerifyingKey = c.Recursion.GetPublicInputOfInstance(run, verifyingKeyPublicInput, i+1)
			sameVerifyingKeyAsNext = verifyingKey == nextVerifyingKey
		}

		if sameVerifyingKeyAsPrev && hornerN0Hash != prevHornerN1Hash {
			mainErr = errors.Join(mainErr, errors.New("horner-n0-hash mismatch"))
		}

		if !sameVerifyingKeyAsPrev != isFirst.IsOne() {
			mainErr = errors.Join(mainErr, errors.New("isFirst is inconsistent with the verifying keys"))
		}

		if !sameVerifyingKeyAsNext != isLast.IsOne() {
			mainErr = errors.Join(mainErr, errors.New("isLast is inconsistent with the verifying keys"))
		}

		if sameVerifyingKeyAsPrev && globalReceived != prevGlobalSent {
			mainErr = errors.Join(mainErr, errors.New("global sent and receive don't match"))
		}

		if !isLPP.IsOne() {
			root := c.Recursion.GetPublicInputOfInstance(run, lppMerkleRootPublicInput+"_0", i)
			merkleRootsGL = append(merkleRootsGL, root)
		}

		if isLPP.IsOne() {
			for j := 0; j < int(nbActualLppPublicInput.Uint64()); j++ {
				var (
					name = fmt.Sprintf("%v_%v", lppMerkleRootPublicInput, j)
					root = c.Recursion.GetPublicInputOfInstance(run, name, i)
				)

				merkleRootsLPP = append(merkleRootsLPP, root)
			}
		}

		prevHornerN1Hash = hornerN1Hash
		prevGlobalSent = globalSent
		allGrandProduct.Mul(&allGrandProduct, &grandProduct)
		allHornerSum.Add(&allHornerSum, &hornerSum)
		allLogDerivativeSum.Add(&allLogDerivativeSum, &logDerivativeSum)
	}

	mapping := map[field.Element]int{}
	for _, v := range merkleRootsGL {
		if _, ok := mapping[v]; !ok {
			mapping[v] = 0
		}

		mapping[v]++
	}

	for _, v := range merkleRootsLPP {
		if _, ok := mapping[v]; !ok {
			mainErr = errors.Join(mainErr, fmt.Errorf("missing public input"))
			continue
		}

		mapping[v]--
	}

	for _, c := range mapping {
		if c != 0 {
			mainErr = errors.Join(mainErr, fmt.Errorf("public input mismatch"))
		}
	}

	return mainErr
}

// cmpWizardIOP compares two compiled IOPs. The function is here to help ensuring
// that all the conglomerated wizard IOPs have the same structure and help
// figuring out inconsistencies if there are.
func cmpWizardIOP(c1, c2 *wizard.CompiledIOP) (diff1, diff2 []string) {

	var (
		stringB1 = &strings.Builder{}
		stringB2 = &strings.Builder{}
	)

	logdata.GenCSV(stringB1, logdata.IncludeAllFilter)(c1)
	logdata.GenCSV(stringB2, logdata.IncludeAllFilter)(c2)

	var (
		c1Formatted = strings.Split(stringB1.String(), "\n")
		c2Formatted = strings.Split(stringB2.String(), "\n")
	)

	diff1, diff2 = utils.SetDiff(c1Formatted, c2Formatted)
	lessFunc := func(a, b string) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		} else {
			return 0
		}
	}

	slices.SortFunc(diff1, lessFunc)
	slices.SortFunc(diff2, lessFunc)

	return diff1, diff2
}
