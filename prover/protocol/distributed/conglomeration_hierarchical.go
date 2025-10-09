package distributed

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
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
	targetNbSegmentPublicInputBase          = "TARGET_NB_SEGMENTS"
	segmentCountLPPPublicInputBase          = "GL_SEGMENT_COUNT"
	segmentCountGLPublicInputBase           = "LPP_SEGMENT_COUNT"
	GeneralMultiSetPublicInputBase          = "GENERAL_MULTI_SET"
	sharedRandomnessMultiSetPublicInputBase = "SHARED_RANDOMNESS_MULTI_SET"
	VkMerkleProofBase                       = "VK_MERKLE_PROOF"
	InitialRandomnessPublicInput            = "INITIAL_RANDOMNESS_PUBLIC_INPUT"
	LogDerivativeSumPublicInput             = "LOG_DERIVATE_SUM_PUBLIC_INPUT"
	GrandProductPublicInput                 = "GRAND_PRODUCT_PUBLIC_INPUT"
	HornerPublicInput                       = "HORNER_FINAL_RES_PUBLIC_INPUT"
	globalHashSentPublicInput               = "GLOBAL_HASH_SENT"
	globalHashReceivedPublicInput           = "GLOBAL_HASH_RECEIVED"
	verifyingKeyPublicInput                 = "VERIFYING_KEY"
	verifyingKey2PublicInput                = "VERIFYING_KEY_2"
	lppMerkleRootPublicInput                = "LPP_COLUMNS_MERKLE_ROOTS"

	// The prerecursion prefix is a prefix to apply to the name of the public
	// inputs to be able to access them in the conglomerated wizard-IOP.
	preRecursionPrefix = "wizard-recursion-0."
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

// VerificationKeyMerkleTree is a Merkle tree storing a list of verification keys
// and it is meant to store the verification keys of all the moduleGL/LPP and
// and of the ConglomerationHierarchical circuit.
type VerificationKeyMerkleTree struct {
	Tree             *smt.Tree
	VerificationKeys [][2]field.Element
}

// BuildVerificationKeyMerkleTree builds the verification key merkle tree.
func BuildVerificationKeyMerkleTree(moduleGL, moduleLPP []*RecursedSegmentCompilation, hierAgg *RecursedSegmentCompilation) VerificationKeyMerkleTree {

	var (
		leaves           = make([]types.Bytes32, 0, len(moduleGL)+len(moduleLPP)+1)
		verificationKeys = make([][2]field.Element, 0, len(moduleGL)+len(moduleLPP)+1)
	)

	appendLeaf := func(comp *wizard.CompiledIOP) {
		var (
			vk0, vk1 = getVerifyingKeyPair(comp)
			leafF    = mimc.HashVec([]field.Element{vk0, vk1})
			leaf     types.Bytes32
		)

		leaf.SetField(leafF)
		leaves = append(leaves, leaf)
		verificationKeys = append(verificationKeys, [2]field.Element{vk0, vk1})
	}

	for _, module := range moduleGL {
		appendLeaf(module.RecursionComp)
	}

	for _, module := range moduleLPP {
		appendLeaf(module.RecursionComp)
	}

	appendLeaf(hierAgg.RecursionComp)

	// padding with zeroes so that the leaves number if a power-of-two
	paddedSize := utils.NextPowerOfTwo(len(leaves))
	for i := len(leaves); i < paddedSize; i++ {
		leaves = append(leaves, types.Bytes32{})
	}

	return VerificationKeyMerkleTree{
		Tree:             smt.BuildComplete(leaves, hashtypes.MiMC),
		VerificationKeys: verificationKeys,
	}
}

// GetVkMerkleProof return the merkle proof of a verification key
func (vmt VerificationKeyMerkleTree) GetVkMerkleProof(t ProofType, moduleIndex int) []field.Element {
	proof := vmt.Tree.MustProve(moduleIndex)
	res := make([]field.Element, len(proof.Siblings))
	for i, sibling := range proof.Siblings {
		res[i].SetBytes(sibling[:])
	}
	return res
}

// CheckMembership checks if a verification key is in the merkle tree.
func checkVkMembership(t ProofType, numModule int, moduleIndex int, vk [2]field.Element, rootF field.Element, proofF []field.Element) error {

	var leafPosition = -1

	switch t {
	// the instance is a conglomeration proof
	case proofTypeConglo:
		leafPosition = 2 * numModule
	case proofTypeLPP:
		leafPosition = moduleIndex + numModule
	case proofTypeGL:
		leafPosition = moduleIndex
	}

	// This part of the loop checks the membership of the VK as a member of
	// the tree using the leafPosition from above.

	var (
		merkleDepth = utils.Log2Ceil(2*numModule + 1)
		root        types.Bytes32
		mProof      = smt.Proof{
			Path:     leafPosition,
			Siblings: make([]types.Bytes32, merkleDepth),
		}
		smtCfg = &smt.Config{HashFunc: hashtypes.MiMC, Depth: merkleDepth}
		leafF  = mimc.HashVec(vk[:])
		leaf   = types.Bytes32{}
	)

	if merkleDepth != len(proofF) {
		panic("merkleDepth != len(proofF)")
	}

	leaf.SetField(leafF)

	for lvl := 0; lvl < merkleDepth; lvl++ {
		mProof.Siblings[lvl].SetField(proofF[lvl])
	}

	root.SetField(rootF)

	if !mProof.Verify(smtCfg, leaf, root) {
		return errors.New("VK is not a member of the tree")
	}

	return nil
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
	Functionals                  []T
	TargetNbSegments             []T
	SegmentCountGL               []T
	SegmentCountLPP              []T
	GeneralMultiSetHash          []T
	SharedRandomnessMultiSetHash []T
	VKeyMerkleRoot               T
	VerifyingKey                 [2]T
	LogDerivativeSum             T
	HornerSum                    T
	GrandProduct                 T
	SharedRandomness             T
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

	c.PublicInputs.TargetNbSegments = declareListOfPiColumns(c.Wiop, 0, targetNbSegmentPublicInputBase, c.ModuleNumber)
	c.PublicInputs.SegmentCountGL = declareListOfPiColumns(c.Wiop, 0, segmentCountGLPublicInputBase, c.ModuleNumber)
	c.PublicInputs.SegmentCountLPP = declareListOfPiColumns(c.Wiop, 0, segmentCountLPPPublicInputBase, c.ModuleNumber)
	c.PublicInputs.GeneralMultiSetHash = declareListOfPiColumns(c.Wiop, 0, GeneralMultiSetPublicInputBase, mimc.MSetHashSize)
	c.PublicInputs.VerifyingKey[0] = declarePiColumn(c.Wiop, verifyingKeyPublicInput)
	c.PublicInputs.VerifyingKey[1] = declarePiColumn(c.Wiop, verifyingKey2PublicInput)
	c.PublicInputs.LogDerivativeSum = declarePiColumn(c.Wiop, LogDerivativeSumPublicInput)
	c.PublicInputs.HornerSum = declarePiColumn(c.Wiop, HornerPublicInput)
	c.PublicInputs.GrandProduct = declarePiColumn(c.Wiop, GrandProductPublicInput)
	c.PublicInputs.SharedRandomness = declarePiColumn(c.Wiop, InitialRandomnessPublicInput)

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

	for instance := 0; instance < aggregationArity; instance++ {
		collectedPIs[instance] = c.collectAllPublicInputsOfInstance(run, instance)
	}

	// This checks that the functional public inputs are correctly conglomerated
	// across all instances.
	for k := range topPIs.Functionals {

		summedUpValue := field.Element{}

		for instance := 0; instance < aggregationArity; instance++ {
			funcPI := collectedPIs[instance].Functionals[k]
			summedUpValue.Add(&summedUpValue, &funcPI)
		}

		if summedUpValue != topPIs.Functionals[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for Functionals at index %d, name=%v", k, c.FunctionalName[k]))
		}
	}

	for k := 0; k < c.ModuleNumber; k++ {

		var (
			sumCountGL  = field.Element{}
			sumCountLPP = field.Element{}
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
		}

		if sumCountGL != topPIs.SegmentCountGL[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for SegmentCountGL for module %d", k))
		}

		if sumCountLPP != topPIs.SegmentCountLPP[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for SegmentCountLPP for module %d", k))
		}
	}

	// This agglomerates the multiset hashes
	for k := 0; k < mimc.MSetHashSize; k++ {

		var (
			generalSum = field.Element{}
			sharedSum  = field.Element{}
		)

		for instance := 0; instance < c.ModuleNumber; instance++ {
			generalSum.Add(&generalSum, &collectedPIs[instance].GeneralMultiSetHash[k])
			sharedSum.Add(&sharedSum, &collectedPIs[instance].SharedRandomnessMultiSetHash[k])
		}

		if generalSum != topPIs.GeneralMultiSetHash[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for generalMultiSetHash for index %d", k))
		}

		if sharedSum != topPIs.SharedRandomnessMultiSetHash[k] {
			err = errors.Join(err, fmt.Errorf("public input mismatch for sharedRandomness for index %d", k))
		}
	}

	// The loop below "aggregate" the public inputs: log-derivative-sum, gd-product,
	// and horner sum of the sub-instances. The aggregation is done by multiplying/summing
	// the values. The results are then compared the top-level public inputs.
	var (
		accGrandProduct = field.One()
		accLogDeriv     = field.Zero()
		accHornerSum    = field.Zero()
	)

	for instance := 0; instance < aggregationArity; instance++ {

		// This agglomerates the horner N0 hash checker, the grand product, the
		// log derivative sum and the horner sum.
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

	if accGrandProduct != topPIs.GrandProduct {
		err = errors.Join(err, fmt.Errorf("public input mismatch for GrandProduct, %v != %v", accGrandProduct.String(), topPIs.GrandProduct.String()))
	}

	if accLogDeriv != topPIs.LogDerivativeSum {
		err = errors.Join(err, fmt.Errorf("public input mismatch for LogDerivativeSum, %v != %v", accLogDeriv.String(), topPIs.LogDerivativeSum.String()))
	}

	if accHornerSum != topPIs.HornerSum {
		err = errors.Join(err, fmt.Errorf("public input mismatch for HornerSum, %v != %v", accHornerSum.String(), topPIs.HornerSum.String()))
	}

	// This loop checks the VK membership in the tree. The merkle leaf position
	// is deduced from the segment count public inputs in the following way;
	//
	// 	- If segment-count-sum of the GL position is one and LPP is zero, then
	//  	the position is the position of the "count=1" GL input.
	//
	//  - If the segment segment-count-sum of the LPP positions is one and GL is
	// 		zero, then the position is the position of the "count=1" LPP input +
	// 		nb-module
	//
	// 	- Otherwise (the total sum is larger than 1), the position is 2*nb-module

	for instance := 0; instance < aggregationArity; instance++ {

		proofType, moduleIndex := findProofTypeAndModule(collectedPIs[instance])

		mProof := make([]field.Element, c.ConglomerationHierarchical.VKeyMTreeDepth())
		for i := range mProof {
			mProof[i] = c.VerificationKeyMerkleProofs[instance][i].GetColAssignmentAt(run, 0)
		}

		vkErr := checkVkMembership(
			proofType,
			c.ModuleNumber,
			moduleIndex,
			collectedPIs[instance].VerifyingKey,
			collectedPIs[instance].VKeyMerkleRoot,
			mProof,
		)

		if vkErr != nil {
			err = errors.Join(err, vkErr)
		}
	}

	if err != nil {
		return err
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
	return declarePiColumnAtRound(comp, 0, name)
}

// declarePiColumn at round declares a column at the requested round to generate
// a public input with the requested name.
func declarePiColumnAtRound(comp *wizard.CompiledIOP, round int, name string) wizard.PublicInput {
	col := comp.InsertProof(round, ifaces.ColID(name+"_PI_COLUMN"), 1)
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
func declareListOfPiColumns(comp *wizard.CompiledIOP, round int, name string, length int) []wizard.PublicInput {
	var cols []wizard.PublicInput
	for i := 0; i < length; i++ {
		cols = append(cols, declarePiColumnAtRound(comp, round, name+"_"+strconv.Itoa(i)))
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

// GetPublicInputList returns a list of public input of the provided name for the
// current WIOP (e.g. not the for the children instance).
//
// @alex: would be interesting to make that a utility function in the wizard
// package because it helps whenever we want to encode stuffs as public inputs.
func GetPublicInputList(run wizard.Runtime, name string, nb int) []field.Element {
	var res []field.Element
	for i := 0; i < nb; i++ {
		res = append(res, run.GetPublicInput(name+"_"+strconv.Itoa(i)))
	}
	return res
}

// getPublicInputListGnark is as [getPublicInputList] but for gnark.
func getPublicInputListGnark(api frontend.API, run wizard.GnarkRuntime, name string, nb int) []frontend.Variable {
	var res []frontend.Variable
	for i := 0; i < nb; i++ {
		res = append(res, run.GetPublicInput(api, name+"_"+strconv.Itoa(i)))
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
		GeneralMultiSetHash: getPublicInputListOfInstance(c.Recursion, run, GeneralMultiSetPublicInputBase, instance, mimc.MSetHashSize),
		VerifyingKey: [2]field.Element{
			c.Recursion.GetPublicInputOfInstance(run, verifyingKeyPublicInput, instance),
			c.Recursion.GetPublicInputOfInstance(run, verifyingKey2PublicInput, instance),
		},
		LogDerivativeSum: c.Recursion.GetPublicInputOfInstance(run, LogDerivativeSumPublicInput, instance),
		HornerSum:        c.Recursion.GetPublicInputOfInstance(run, HornerPublicInput, instance),
		GrandProduct:     c.Recursion.GetPublicInputOfInstance(run, GrandProductPublicInput, instance),
		SharedRandomness: c.Recursion.GetPublicInputOfInstance(run, InitialRandomnessPublicInput, instance),
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
		TargetNbSegments:    GetPublicInputList(run, targetNbSegmentPublicInputBase, c.ModuleNumber),
		SegmentCountGL:      GetPublicInputList(run, segmentCountGLPublicInputBase, c.ModuleNumber),
		SegmentCountLPP:     GetPublicInputList(run, segmentCountLPPPublicInputBase, c.ModuleNumber),
		GeneralMultiSetHash: GetPublicInputList(run, GeneralMultiSetPublicInputBase, mimc.MSetHashSize),
		VerifyingKey: [2]field.Element{
			run.GetPublicInput(verifyingKeyPublicInput),
			run.GetPublicInput(verifyingKey2PublicInput),
		},
		LogDerivativeSum: run.GetPublicInput(LogDerivativeSumPublicInput),
		HornerSum:        run.GetPublicInput(HornerPublicInput),
		GrandProduct:     run.GetPublicInput(GrandProductPublicInput),
		SharedRandomness: run.GetPublicInput(InitialRandomnessPublicInput),
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

// findProofTypeAndModule returns the proofType and the module index of the
// provided instance given collected public inputs of the instances.
func findProofTypeAndModule(instance LimitlessPublicInput[field.Element]) (ProofType, int) {

	var (
		sumGL, sumLPP = 0, 0
		moduleIndex   = -1 // can't be -1 at the end of the "mod" loop.
		moduleNumber  = len(instance.SegmentCountGL)
	)

	for mod := 0; mod < moduleNumber; mod++ {

		var (
			segmentCountGL  = instance.SegmentCountGL[mod]
			segmentCountLPP = instance.SegmentCountLPP[mod]
		)

		sumGL += int(segmentCountGL.Uint64())
		sumLPP += int(segmentCountLPP.Uint64())

		if !segmentCountGL.IsZero() || !segmentCountLPP.IsZero() {
			moduleIndex = mod
		}
	}

	switch {
	case sumGL+sumLPP > 1:
		return proofTypeConglo, 0
	case sumGL >= 1:
		return proofTypeGL, moduleIndex
	case sumLPP >= 1:
		return proofTypeLPP, moduleIndex
	}

	panic("unreachable")
}

// getVerifyingKeyPair extracts the verifyingKeys from the compiled IOP.
func getVerifyingKeyPair(wiop *wizard.CompiledIOP) (vkGL, vkLPP field.Element) {
	return wiop.ExtraData[verifyingKeyPublicInput].(field.Element),
		wiop.ExtraData[verifyingKey2PublicInput].(field.Element)
}
