package distributed

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type ProofType int

const (

	// types enum for the proof types
	proofTypeLPP ProofType = iota
	proofTypeGL
	proofTypeConglo

	// aggregationArity is the arity of the aggregation circuit
	aggregationArity = 2

	// name of the public inputs
	targetNbSegmentPublicInputBase    = "TARGET_NB_SEGMENTS"
	segmentCountLPPPublicInputBase    = "GL_SEGMENT_COUNT"
	segmentCountGLPublicInputBase     = "LPP_SEGMENT_COUNT"
	lppCommitmentMSetPublicInputBase  = "LPP_COMMITMENT_SET"
	hornerN0HashCheckerPublicInput    = "HORNER_N0_HASH_CHECKER"
	GlobalSentReceivedMSetPublicInput = "GLOBAL_SENT_RECEIVED_MSET"
	VkMerkleProofBase                 = "VK_MERKLE_PROOF"
	InitialRandomnessPublicInput      = "INITIAL_RANDOMNESS_PUBLIC_INPUT"
	LogDerivativeSumPublicInput       = "LOG_DERIVATE_SUM_PUBLIC_INPUT"
	GrandProductPublicInput           = "GRAND_PRODUCT_PUBLIC_INPUT"
	HornerPublicInput                 = "HORNER_FINAL_RES_PUBLIC_INPUT"
	verifyingKeyPublicInput           = "VERIFYING_KEY"
	verifyingKey2PublicInput          = "VERIFYING_KEY_2"
	lppMerkleRootPublicInput          = "LPP_COLUMNS_MERKLE_ROOTS"
)

// ConglomerationCompilation holds the compilation context of the hierarchical
// conglomeration.
type ConglomerationHierarchical struct {
	// ModuleNumber gives the number of modules of the distributed prover
	ModuleNumber int
	// FunctionalName lists the name of the functional public-inputs
	FunctionalName []string
	// Wiop is the compiled IOP of the conglomeration wizard.
	Wiop *wizard.CompiledIOP
	// Recursion is the recursion context used to compile the conglomeration
	// proof.
	Recursion *recursion.Recursion
	// PublicInputs stores the public inputs of the conglomeration proof.
	PublicInputs LimitlessPublicInput[wizard.PublicInput]

	// VerificationKeyMerkleProofs is the list of the verification keys proving
	// the membership of the verifying keys of the instances inside the
	// VerificationKeyMerkleTree. Each merkle proof is structured as a list of
	// D columns if size 1 where D is the depth of the merkle tree.
	VerificationKeyMerkleProofs [][]ifaces.Column
}

// ConglomerationHierarchicalVerifierAction implements the [wizard.VerifierAction]
// interface for the conglomeration proof. It checks the consistency of the
// public inputs with the children instance's public inputs.
type ConglomerationHierarchicalVerifierAction struct {
	ConglomerationHierarchical
}

// LimitlessPublicInput stores the columns totalling the
// public inputs of a conglomeration node.
type LimitlessPublicInput[T any] struct {
	Functionals            []T
	TargetNbSegments       []T
	SegmentCountGL         []T
	SegmentCountLPP        []T
	LppCommitmentMSetGL    []T
	LppCommitmentMSetLPP   []T
	SegmentIndexChecker    []T
	VKeyMerkleRoot         T
	VerifyingKey           [2]T
	LogDerivativeSum       T
	HornerSum              T
	GrandProduct           T
	SharedRandomness       T
	HornerN0HashChecker    T
	GlobalSentReceivedMSet []T
}

// Compile compiles the conglomeration proof. The function first checks if the
// public inputs are compatible and then compiles the conglomeration proof.
func (c *ConglomerationHierarchical) Compile(comp *wizard.CompiledIOP, moduleMod *wizard.CompiledIOP) {

	c.Recursion = recursion.DefineRecursionOf(comp, moduleMod, recursion.Parameters{
		Name:                   "conglomeration",
		WithoutGkr:             true,
		MaxNumProof:            2,
		WithExternalHasherOpts: true,
	})

	c.Wiop = comp

	for _, name := range c.FunctionalName {
		c.PublicInputs.Functionals = append(c.PublicInputs.Functionals, declarePiColumn(c.Wiop, name))
	}

	c.PublicInputs.TargetNbSegments = declareListOfPiColumns(c.Wiop, targetNbSegmentPublicInputBase, c.ModuleNumber)
	c.PublicInputs.SegmentCountGL = declareListOfPiColumns(c.Wiop, segmentCountGLPublicInputBase, c.ModuleNumber)
	c.PublicInputs.SegmentCountLPP = declareListOfPiColumns(c.Wiop, segmentCountLPPPublicInputBase, c.ModuleNumber)
	c.PublicInputs.LppCommitmentMSetGL = declareListOfPiColumns(c.Wiop, lppCommitmentMSetPublicInputBase, mimc.MSetHashSize)
	c.PublicInputs.VerifyingKey[0] = declarePiColumn(c.Wiop, verifyingKeyPublicInput)
	c.PublicInputs.VerifyingKey[1] = declarePiColumn(c.Wiop, verifyingKey2PublicInput)
	c.PublicInputs.LogDerivativeSum = declarePiColumn(c.Wiop, LogDerivativeSumPublicInput)
	c.PublicInputs.HornerSum = declarePiColumn(c.Wiop, HornerPublicInput)
	c.PublicInputs.GrandProduct = declarePiColumn(c.Wiop, GrandProductPublicInput)
	c.PublicInputs.SharedRandomness = declarePiColumn(c.Wiop, InitialRandomnessPublicInput)
	c.PublicInputs.HornerN0HashChecker = declarePiColumn(c.Wiop, hornerN0HashCheckerPublicInput)
	c.PublicInputs.GlobalSentReceivedMSet = declareListOfPiColumns(c.Wiop, GlobalSentReceivedMSetPublicInput, mimc.MSetHashSize)

	// vkMerkleTreeDepth is the depth of the verification key merkle tree
	vkMerkleTreeDepth := c.VKeyMTreeDepth()
	c.VerificationKeyMerkleProofs = make([][]ifaces.Column, c.ModuleNumber)
	for i := 0; i < c.ModuleNumber; i++ {
		for j := 0; j < vkMerkleTreeDepth; j++ {
			col := comp.InsertProof(0, ifaces.ColID(fmt.Sprintf("vkMerkleProof_%d_%d", i, j)), 1)
			c.VerificationKeyMerkleProofs[i] = append(c.VerificationKeyMerkleProofs[i], col)
		}
	}

}

// Assign assigns the public inputs for the conglomeration proof
func (c *ConglomerationHierarchical) Assign(
	run *wizard.ProverRuntime,
	proofs []recursion.Witness,
) {

	// This runs the recursion system. Expectedly, the filling input is never
	// used because this is pairwise aggregation and we always pass exactly pass
	// exactly 2 inputs.
	c.Recursion.Assign(run, proofs, &proofs[0])

	// Now, it remains to assign the public inputs for the conglomeration proof.
	var (
		collectedPIs = [aggregationArity]LimitlessPublicInput[field.Element]{}
		sumCountGLs  = []field.Element{}
		sumCountLPPs = []field.Element{}
	)

	for instance := 0; instance < c.ModuleNumber; instance++ {
		collectedPIs[instance] = c.collectAllPublicInputsOfInstance(run, instance)
	}

	for k := 0; k < c.ModuleNumber; k++ {

		var sumCountGL, sumCountLPP field.Element

		for instance := 0; instance < aggregationArity; instance++ {
			// This agglomerates the segment count for the GL and the LPPs modules. There
			// is one GL and one LPP counter for each module that's why we do them in the
			sumCountGL.Add(&sumCountGL, &collectedPIs[instance].SegmentCountGL[k])
			sumCountLPP.Add(&sumCountLPP, &collectedPIs[instance].SegmentCountLPP[k])
		}

		sumCountGLs = append(sumCountGLs, sumCountGL)
		sumCountLPPs = append(sumCountLPPs, sumCountLPP)
	}

	assignListOfPiColumns(run, targetNbSegmentPublicInputBase, collectedPIs[0].TargetNbSegments)
	assignListOfPiColumns(run, segmentCountGLPublicInputBase, sumCountGLs)
	assignListOfPiColumns(run, segmentCountLPPPublicInputBase, sumCountLPPs)
}

// Run implements the [wizard.VerifierAction] for the
// ConglomerationHierarchicalVerifierAction.
func (c *ConglomerationHierarchicalVerifierAction) Run(run wizard.Runtime) error {

	var (
		err          error
		collectedPIs = [aggregationArity]LimitlessPublicInput[field.Element]{}
		topPIs       = c.collectAllPublicInputs(run)
	)

	for instance := 0; instance < c.ModuleNumber; instance++ {
		collectedPIs[instance] = c.collectAllPublicInputsOfInstance(run, instance)
	}

	// This checks that the functional public inputs are correctly conglomerated
	// across all instances.
	for k := range topPIs.Functionals {

		summedUpValue := field.Element{}

		for instance := 0; instance < c.ModuleNumber; instance++ {
			funcPI := collectedPIs[instance].Functionals[k]
			summedUpValue.Add(&summedUpValue, &funcPI)
		}

		if summedUpValue != topPIs.Functionals[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for Functionals at index %d, name=%v", k, c.FunctionalName[k]))
		}
	}

	for k := 0; k < c.ModuleNumber; k++ {

		var (
			accSegmentIndexChecker = field.One()
			sumCountGL             = field.Element{}
			sumCountLPP            = field.Element{}
		)

		for instance := 0; instance < aggregationArity; instance++ {

			// This checks that the TargetNbSegments public inputs are the same for all
			// the children instances and the current node.
			if collectedPIs[instance].TargetNbSegments[k] != topPIs.TargetNbSegments[k] {
				err = errors.Join(err, fmt.Errorf("public input mismatch for TargetNbSegments at instance %d", instance))
			}

			// This agglomerates the segment count for the GL and the LPPs modules. There
			// is one GL and one LPP counter for each module that's why we do them in the
			sumCountGL.Add(&sumCountGL, &collectedPIs[instance].SegmentCountGL[k])
			sumCountLPP.Add(&sumCountLPP, &collectedPIs[instance].SegmentCountLPP[k])

			// This agglomerates the segment index checkers and the horner N0
			// hash checker.
			accSegmentIndexChecker.Mul(&accSegmentIndexChecker, &collectedPIs[instance].SegmentIndexChecker[k])
		}

		if sumCountGL != topPIs.SegmentCountGL[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for SegmentCountGL for module %d", k))
		}

		if sumCountLPP != topPIs.SegmentCountLPP[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for SegmentCountLPP for module %d", k))
		}

		if accSegmentIndexChecker != topPIs.SegmentIndexChecker[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for SegmentIndexChecker for module %d", k))
		}
	}

	// This agglomerates the multiset hashes
	for k := 0; k < mimc.MSetHashSize; k++ {

		var (
			sumHashGL  = field.Element{}
			sumHashLPP = field.Element{}
			sumGlobal  = field.Element{}
		)

		for instance := 0; instance < c.ModuleNumber; instance++ {
			sumHashGL.Add(&sumHashGL, &collectedPIs[instance].LppCommitmentMSetGL[k])
			sumHashLPP.Add(&sumHashLPP, &collectedPIs[instance].LppCommitmentMSetLPP[k])
			sumGlobal.Add(&sumGlobal, &collectedPIs[instance].GlobalSentReceivedMSet[k])
		}

		if sumHashGL != topPIs.LppCommitmentMSetGL[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for LppCommitmentMSetGL for index %d", k))
		}

		if sumHashLPP != topPIs.LppCommitmentMSetLPP[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for LppCommitmentMSetLPP for index %d", k))
		}

		if sumGlobal != topPIs.GlobalSentReceivedMSet[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for GlobalSentReceivedMSet for index %d", k))
		}
	}

	var (
		accN0HashChecker = field.One()
		accGrandProduct  = field.One()
		accLogDeriv      = field.Zero()
		accHornerSum     = field.Zero()
	)

	for instance := 0; instance < c.ModuleNumber; instance++ {

		// This agglomerates the horner N0 hash checker, the grand product, the
		// log derivative sum and the horner sum.
		accN0HashChecker.Mul(&accN0HashChecker, &collectedPIs[instance].HornerN0HashChecker)
		accGrandProduct.Mul(&accGrandProduct, &collectedPIs[instance].GrandProduct)
		accLogDeriv.Add(&accLogDeriv, &collectedPIs[instance].LogDerivativeSum)
		accHornerSum.Add(&accHornerSum, &collectedPIs[instance].HornerSum)

		if !collectedPIs[instance].SharedRandomness.IsZero() && collectedPIs[instance].SharedRandomness != topPIs.SharedRandomness {
			err = errors.Join(err, fmt.Errorf("public input mismatch for SharedRandomness for instance %d", instance))
		}

		if collectedPIs[instance].VKeyMerkleRoot == topPIs.VKeyMerkleRoot {
			err = errors.Join(err, fmt.Errorf("public input mismatch for VKeyMerkleRoot for instance %d", instance))
		}
	}

	if accN0HashChecker != topPIs.HornerN0HashChecker {
		err = errors.Join(err, fmt.Errorf("public input mismatch for HornerN0HashChecker, %v != %v", accN0HashChecker.String(), topPIs.HornerN0HashChecker.String()))
	}

	if accGrandProduct != topPIs.GrandProduct {
		err = errors.Join(err, fmt.Errorf("public input mismatch for GrandProduct, %v != %v", accGrandProduct.String(), topPIs.GrandProduct.String()))
	}

	if accLogDeriv != topPIs.LogDerivativeSum {
		err = errors.Join(err, fmt.Errorf("public input mismatch for LogDerivativeSum, %v != %v", accLogDeriv.String(), topPIs.LogDerivativeSum.String()))
	}

	if accHornerSum != topPIs.HornerSum {
		err = errors.Join(err, fmt.Errorf("public input mismatch for HornerSum, %v != %v", accHornerSum.String(), topPIs.HornerSum.String()))
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction] interface.
func (c *ConglomerationHierarchicalVerifierAction) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	panic("unimplemented")
}

// declarePi declares a column with the requested name as proof column and length
// one and also declare a public input from that column with the same provided
// name.
func declarePiColumn(comp *wizard.CompiledIOP, name string) wizard.PublicInput {
	col := comp.InsertProof(0, ifaces.ColID(name+"_PI_COLUMN"), 1)
	return comp.InsertPublicInput(name, accessors.NewFromPublicColumn(col, 0))
}

// assignPiColumn assigns the column of a public input with the requested name
// to the provided column.
func assignPiColumn(run *wizard.ProverRuntime, name string, val field.Element) {
	run.AssignColumn(
		ifaces.ColID(name+"_PI_COLUMN"),
		smartvectors.NewConstant(val, 1),
	)
}

// declareListOfPiColumns declares a list of columns with the requested name as
// proof columns and length provided.
func declareListOfPiColumns(comp *wizard.CompiledIOP, name string, length int) []wizard.PublicInput {
	var cols []wizard.PublicInput
	for i := 0; i < length; i++ {
		cols = append(cols, declarePiColumn(comp, name+"_"+strconv.Itoa(i)))
	}
	return cols
}

// declareListOfConstantPi declares a list of public inputs as constant values.
// This is useful to create dummy public inputs making the aggregation process
// simpler.
func declareListOfConstantPi(comp *wizard.CompiledIOP, name string, values []field.Element) []wizard.PublicInput {
	res := make([]wizard.PublicInput, len(values))
	for i, val := range values {
		pub := comp.InsertPublicInput(name+"_"+strconv.Itoa(i), accessors.NewConstant(val))
		res = append(res, pub)
	}
	return res
}

// assignListOfPiColumns assigns a list of columns with the requested name using
// the provided list of values.
func assignListOfPiColumns(run *wizard.ProverRuntime, name string, values []field.Element) {
	for i, val := range values {
		assignPiColumn(run, name+"_"+strconv.Itoa(i), val)
	}
}

// getPublicInputListOfInstance returns a list of public inputs of the provided
// name and instance.
func getPublicInputListOfInstance(rec *recursion.Recursion, run wizard.Runtime, name string, instance int, nb int) []field.Element {
	var res []field.Element
	for i := 0; i < nb; i++ {
		res = append(res, rec.GetPublicInputOfInstance(run, name, instance))
	}
	return res
}

// getPublicInputList returns a list of public input of the provided name for the
// current WIOP (e.g. not the for the children instance).
func getPublicInputList(run wizard.Runtime, name string, nb int) []field.Element {
	var res []field.Element
	for i := 0; i < nb; i++ {
		res = append(res, run.GetPublicInput(name+"_"+strconv.Itoa(i)))
	}
	return res
}

// collectAllPublicInputsOfInstance returns a structured object representing
// the public inputs of the given instance.
func (c ConglomerationHierarchical) collectAllPublicInputsOfInstance(run wizard.Runtime, instance int) LimitlessPublicInput[field.Element] {

	res := LimitlessPublicInput[field.Element]{
		TargetNbSegments:    getPublicInputListOfInstance(c.Recursion, run, targetNbSegmentPublicInputBase, instance, c.ModuleNumber),
		SegmentCountGL:      getPublicInputListOfInstance(c.Recursion, run, segmentCountGLPublicInputBase, instance, c.ModuleNumber),
		SegmentCountLPP:     getPublicInputListOfInstance(c.Recursion, run, segmentCountLPPPublicInputBase, instance, c.ModuleNumber),
		LppCommitmentMSetGL: getPublicInputListOfInstance(c.Recursion, run, lppCommitmentMSetPublicInputBase, instance, mimc.MSetHashSize),
		VerifyingKey: [2]field.Element{
			c.Recursion.GetPublicInputOfInstance(run, verifyingKeyPublicInput, instance),
			c.Recursion.GetPublicInputOfInstance(run, verifyingKey2PublicInput, instance),
		},
		LogDerivativeSum:       c.Recursion.GetPublicInputOfInstance(run, LogDerivativeSumPublicInput, instance),
		HornerSum:              c.Recursion.GetPublicInputOfInstance(run, HornerPublicInput, instance),
		GrandProduct:           c.Recursion.GetPublicInputOfInstance(run, GrandProductPublicInput, instance),
		SharedRandomness:       c.Recursion.GetPublicInputOfInstance(run, InitialRandomnessPublicInput, instance),
		HornerN0HashChecker:    c.Recursion.GetPublicInputOfInstance(run, hornerN0HashCheckerPublicInput, instance),
		GlobalSentReceivedMSet: getPublicInputListOfInstance(c.Recursion, run, GlobalSentReceivedMSetPublicInput, instance, mimc.MSetHashSize),
	}

	for _, name := range c.FunctionalName {
		res.Functionals = append(res.Functionals, c.Recursion.GetPublicInputOfInstance(run, name, instance))
	}

	return res
}

// collectAllPublicInputs returns a structured object representing the public
// inputs of all the instances.
func (c ConglomerationHierarchical) collectAllPublicInputs(run wizard.Runtime) LimitlessPublicInput[field.Element] {

	res := LimitlessPublicInput[field.Element]{
		TargetNbSegments:    getPublicInputList(run, targetNbSegmentPublicInputBase, c.ModuleNumber),
		SegmentCountGL:      getPublicInputList(run, segmentCountGLPublicInputBase, c.ModuleNumber),
		SegmentCountLPP:     getPublicInputList(run, segmentCountLPPPublicInputBase, c.ModuleNumber),
		LppCommitmentMSetGL: getPublicInputList(run, lppCommitmentMSetPublicInputBase, mimc.MSetHashSize),
		VerifyingKey: [2]field.Element{
			run.GetPublicInput(verifyingKeyPublicInput),
			run.GetPublicInput(verifyingKey2PublicInput),
		},
		LogDerivativeSum:       run.GetPublicInput(LogDerivativeSumPublicInput),
		HornerSum:              run.GetPublicInput(HornerPublicInput),
		GrandProduct:           run.GetPublicInput(GrandProductPublicInput),
		SharedRandomness:       run.GetPublicInput(InitialRandomnessPublicInput),
		HornerN0HashChecker:    run.GetPublicInput(hornerN0HashCheckerPublicInput),
		GlobalSentReceivedMSet: getPublicInputList(run, GlobalSentReceivedMSetPublicInput, mimc.MSetHashSize),
	}

	for _, name := range c.FunctionalName {
		res.Functionals = append(res.Functionals, run.GetPublicInput(name))
	}

	return res
}

// VKeyMTreeDepth returns the depth of the verification key merkle tree.
func (c ConglomerationHierarchical) VKeyMTreeDepth() int {
	return utils.Log2Ceil(2*c.ModuleNumber + 1)
}
