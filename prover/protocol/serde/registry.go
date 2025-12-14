package serde

import (
	"reflect"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitchsplit"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/expr_handle"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/functionals"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/merkle"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/selector"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/serde/core"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/bls"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecarith"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecpair"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/importpad"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	gen_acc "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/acc_module"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/base_conversion"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing/dedicated/spaghettifier"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/sha2"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/modexp"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/p256verify"

	cmimc "github.com/consensys/linea-monorepo/prover/crypto/mimc"
	dmimc "github.com/consensys/linea-monorepo/prover/protocol/dedicated/mimc"
	ded "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing/dedicated"
)

func init() {
	// This registers all the types that we may need to deserialize
	// as interface implementations. Cases of interfaces that are relevant
	// are:
	//
	// 		- symbolic.Operator
	//      - symbolic.Metadata
	//      - ifaces.Column
	//  	- ifaces.Query
	//

	// The primitive types. Useful when serializing "any" objects
	RegisterImplementation(bool(false))
	RegisterImplementation(int8(0))
	RegisterImplementation(int16(0))
	RegisterImplementation(int32(0))
	RegisterImplementation(int64(0))
	RegisterImplementation(int(0))
	RegisterImplementation(uint8(0))
	RegisterImplementation(uint16(0))
	RegisterImplementation(uint32(0))
	RegisterImplementation(uint64(0))
	RegisterImplementation(uint(0))
	RegisterImplementation(string(""))

	// Interfaces
	RegisterImplementation(ifaces.ColID(""))
	RegisterImplementation(ifaces.QueryID(""))

	// Coins and Columns
	RegisterImplementation(column.Natural{})
	RegisterImplementation(column.Shifted{})
	RegisterImplementation(coin.Name(""))
	RegisterImplementation(coin.Info{})
	RegisterImplementation(column.FakeColumn{})

	// Verifier columns
	RegisterImplementation(verifiercol.ConstCol{})
	RegisterImplementation(verifiercol.FromYs{})
	RegisterImplementation(verifiercol.FromAccessors{})
	RegisterImplementation(verifiercol.ExpandedProofOrVerifyingKeyColWithZero{})
	RegisterImplementation(verifiercol.RepeatedAccessor{})

	// Queries
	RegisterImplementation(query.FixedPermutation{})
	RegisterImplementation(query.GlobalConstraint{})
	RegisterImplementation(query.Inclusion{})
	RegisterImplementation(query.InnerProduct{})
	RegisterImplementation(query.LocalConstraint{})
	RegisterImplementation(query.LocalOpening{})
	RegisterImplementation(query.MiMC{})
	RegisterImplementation(query.Permutation{})
	RegisterImplementation(query.Range{})
	RegisterImplementation(query.UnivariateEval{})
	RegisterImplementation(query.Projection{})
	RegisterImplementation(query.PlonkInWizard{})
	RegisterImplementation(query.LocalOpening{})
	RegisterImplementation(query.LogDerivativeSum{})
	RegisterImplementation(query.GrandProduct{})
	RegisterImplementation(query.Horner{})
	RegisterImplementation(query.HornerParams{})
	RegisterImplementation(query.LocalOpeningParams{})
	RegisterImplementation(query.UnivariateEvalParams{})
	RegisterImplementation(query.GrandProductParams{})
	RegisterImplementation(query.InnerProductParams{})
	RegisterImplementation(query.LogDerivSumParams{})

	// Symbolic
	RegisterImplementation(symbolic.Variable{})
	RegisterImplementation(symbolic.Constant{})
	RegisterImplementation(symbolic.Product{})
	RegisterImplementation(symbolic.LinComb{})
	RegisterImplementation(symbolic.PolyEval{})
	RegisterImplementation(symbolic.StringVar(""))

	// Accessors
	RegisterImplementation(accessors.FromCoinAccessor{})
	RegisterImplementation(accessors.FromConstAccessor{})
	RegisterImplementation(accessors.FromExprAccessor{})
	RegisterImplementation(accessors.FromIntVecCoinPositionAccessor{})
	RegisterImplementation(accessors.FromLocalOpeningYAccessor{})
	RegisterImplementation(accessors.FromPublicColumn{})
	RegisterImplementation(accessors.FromUnivXAccessor{})
	RegisterImplementation(accessors.FromLogDerivSumAccessor{})
	RegisterImplementation(accessors.FromGrandProductAccessor{})
	RegisterImplementation(accessors.FromHornerAccessorFinalValue{})

	// Variables
	RegisterImplementation(variables.X{})
	RegisterImplementation(variables.PeriodicSample{})

	// Circuit implementations
	RegisterImplementation(ecdsa.MultiEcRecoverCircuit{})
	RegisterImplementation(modexp.ModExpCircuit{})
	RegisterImplementation(ecarith.MultiECAddCircuit{})
	RegisterImplementation(ecarith.MultiECMulCircuit{})
	RegisterImplementation(ecpair.MultiG2GroupcheckCircuit{})
	RegisterImplementation(ecpair.MultiMillerLoopMulCircuit{})
	RegisterImplementation(ecpair.MultiMillerLoopFinalExpCircuit{})
	RegisterImplementation(sha2.SHA2Circuit{})

	// BLS circuit implementations - register concrete generic instantiations
	// These types are needed for serializing BLS circuits with their specific type parameters
	// Register zero-value instances so the serialization system can identify the types
	RegisterImplementation(bls.MultiAddCircuitG1{})
	RegisterImplementation(bls.MultiMulCircuitG1{})
	RegisterImplementation(bls.MultiMapCircuitG1{})
	RegisterImplementation(bls.MultiAddCircuitG2{})
	RegisterImplementation(bls.MultiMulCircuitG2{})
	RegisterImplementation(bls.MultiMapCircuitG2{})
	RegisterImplementation(bls.MultiCheckableG1NonGroup{})
	RegisterImplementation(bls.MultiCheckableG1NonCurve{})
	RegisterImplementation(bls.MultiCheckableG2NonGroup{})
	RegisterImplementation(bls.MultiCheckableG2NonCurve{})
	RegisterImplementation(bls.MultiMillerLoopMulCircuit{})
	RegisterImplementation(bls.MultiMillerLoopFinalExpCircuit{})
	RegisterImplementation(bls.MultiPointEvalCircuit{})
	RegisterImplementation(bls.MultiPointEvalFailureCircuit{})

	// P256 verify circuit implementations
	RegisterImplementation(p256verify.MultiP256VerifyInstanceCircuit{})

	// Dedicated and common types
	RegisterImplementation(byte32cmp.MultiLimbCmp{})
	RegisterImplementation(byte32cmp.OneLimbCmpCtx{})
	RegisterImplementation(byte32cmp.DecompositionCtx{})
	RegisterImplementation(dedicated.IsZeroCtx{})
	RegisterImplementation(common.HashingCtx{})

	// Prover actions (added to fix missing concrete type warnings)
	RegisterImplementation(byte32cmp.Bytes32CmpProverAction{})
	RegisterImplementation(bigrange.BigRangeProverAction{})
	RegisterImplementation(ded.AssignPIPProverAction{})
	RegisterImplementation(keccak.ShakiraProverAction{})
	RegisterImplementation(vortex.ColumnAssignmentProverAction{})
	RegisterImplementation(vortex.LinearCombinationComputationProverAction{})

	// Smartvectors
	RegisterImplementation(smartvectors.Regular{})
	RegisterImplementation(smartvectors.PaddedCircularWindow{})
	RegisterImplementation(smartvectors.Constant{})
	RegisterImplementation(smartvectors.Pooled{})

	RegisterImplementation(stitchsplit.ProveRoundProverAction{})
	RegisterImplementation(stitchsplit.AssignLocalPointProverAction{})
	RegisterImplementation(stitchsplit.StitchColumnsProverAction{})
	RegisterImplementation(stitchsplit.StitchSubColumnsProverAction{})
	RegisterImplementation(stitchsplit.QueryVerifierAction{})
	RegisterImplementation(stitchsplit.SplitProverAction{})

	RegisterImplementation(cleanup.CleanupProverAction{})

	RegisterImplementation(dmimc.LinearHashProverAction{})
	RegisterImplementation(merkle.MerkleProofProverAction{})

	RegisterImplementation(univariates.NaturalizeProverAction{})
	RegisterImplementation(univariates.NaturalizeVerifierAction{})

	RegisterImplementation(gen_acc.GenericDataAccumulator{})
	RegisterImplementation(gen_acc.GenericInfoAccumulator{})

	RegisterImplementation(keccak.KeccakSingleProvider{})

	RegisterImplementation(packing.Packing{})

	RegisterImplementation(ded.LengthConsistencyCtx{})
	RegisterImplementation(ded.AccumulateUpToMaxCtx{})

	RegisterImplementation(spaghettifier.Spaghettification{})

	RegisterImplementation(importpad.Sha2Padder{})
	RegisterImplementation(importpad.MimcPadder{})
	RegisterImplementation(importpad.KeccakPadder{})
	RegisterImplementation(importpad.Importation{})
	RegisterImplementation(importpad.Importation{})

	RegisterImplementation(base_conversion.HashBaseConversion{})
	RegisterImplementation(base_conversion.BlockBaseConversion{})
	RegisterImplementation(base_conversion.DecompositionCtx{})

	RegisterImplementation(cleanup.CleanupProverAction{})
	RegisterImplementation(dummy.DummyVerifierAction{})
	RegisterImplementation(dummy.DummyProverAction{})

	RegisterImplementation(globalcs.EvaluationProver{})
	RegisterImplementation(globalcs.EvaluationVerifier{})
	RegisterImplementation(globalcs.QuotientCtx{})

	RegisterImplementation(horner.AssignHornerCtx{})
	RegisterImplementation(horner.AssignHornerIP{})
	RegisterImplementation(horner.AssignHornerQuery{})
	RegisterImplementation(horner.CheckHornerQuery{})
	RegisterImplementation(horner.CheckHornerResult{})

	RegisterImplementation(innerproduct.ProverTask{})
	RegisterImplementation(innerproduct.VerifierForSize{})

	RegisterImplementation(logderivativesum.AssignLogDerivativeSumProverAction{})
	RegisterImplementation(logderivativesum.CheckLogDerivativeSumMustBeZero{})
	RegisterImplementation(logderivativesum.ProverTaskAtRound{})
	RegisterImplementation(logderivativesum.FinalEvaluationCheck{})

	RegisterImplementation(mimc.MimcContext{})
	RegisterImplementation(mpts.QuotientAccumulation{})
	RegisterImplementation(mpts.RandomPointEvaluation{})
	RegisterImplementation(mpts.ShadowRowProverAction{})
	RegisterImplementation(mpts.VerifierAction{})

	RegisterImplementation(permutation.ProverTaskAtRound{})
	RegisterImplementation(permutation.AssignPermutationGrandProduct{})
	RegisterImplementation(permutation.FinalProductCheck{})
	RegisterImplementation(permutation.CheckGrandProductIsOne{})

	RegisterImplementation(plonkinwizard.AssignSelOpening{})
	RegisterImplementation(plonkinwizard.CheckActivatorAndMask{})
	RegisterImplementation(plonkinwizard.CircAssignment{})

	RegisterImplementation(recursion.RecursionCircuit{})
	RegisterImplementation(recursion.AssignVortexOpenedCols{})
	RegisterImplementation(recursion.AssignVortexUAlpha{})
	RegisterImplementation(recursion.ConsistencyCheck{})

	RegisterImplementation(selfrecursion.ColSelectionProverAction{})
	RegisterImplementation(selfrecursion.CollapsingProverAction{})
	RegisterImplementation(selfrecursion.CollapsingVerifierAction{})
	RegisterImplementation(selfrecursion.ConsistencyYsUalphaVerifierAction{})
	RegisterImplementation(selfrecursion.FoldPhaseProverAction{})
	RegisterImplementation(selfrecursion.FoldPhaseVerifierAction{})
	RegisterImplementation(selfrecursion.LinearHashMerkleProverAction{})
	RegisterImplementation(selfrecursion.PreimageLimbsProverAction{})

	RegisterImplementation(stitchsplit.AssignLocalPointProverAction{})
	RegisterImplementation(stitchsplit.ProveRoundProverAction{})
	RegisterImplementation(stitchsplit.QueryVerifierAction{})
	RegisterImplementation(stitchsplit.SplitProverAction{})
	RegisterImplementation(stitchsplit.StitchColumnsProverAction{})
	RegisterImplementation(stitchsplit.StitchSubColumnsProverAction{})

	RegisterImplementation(univariates.NaturalizeProverAction{})
	RegisterImplementation(univariates.NaturalizeVerifierAction{})

	RegisterImplementation(vortex.Ctx{})
	RegisterImplementation(vortex.ColumnAssignmentProverAction{})
	RegisterImplementation(vortex.OpenSelectedColumnsProverAction{})
	RegisterImplementation(vortex.LinearCombinationComputationProverAction{})
	RegisterImplementation(vortex.VortexVerifierAction{})
	RegisterImplementation(vortex.ExplicitPolynomialEval{})
	RegisterImplementation(vortex.ShadowRowProverAction{})
	RegisterImplementation(vortex.ReassignPrecomputedRootAction{})

	RegisterImplementation(functionals.CoeffEvalProverAction{})
	RegisterImplementation(functionals.InterpolationProverAction{})
	RegisterImplementation(functionals.EvalBivariateProverAction{})
	RegisterImplementation(functionals.FoldProverAction{})
	RegisterImplementation(functionals.FoldOuterProverAction{})
	RegisterImplementation(functionals.FoldOuterVerifierAction{})
	RegisterImplementation(functionals.FoldVerifierAction{})
	RegisterImplementation(functionals.XYPow1MinNAccessor{})

	RegisterImplementation(reedsolomon.ReedSolomonProverAction{})
	RegisterImplementation(reedsolomon.ReedSolomonVerifierAction{})

	RegisterImplementation(column.FakeColumn{})

	RegisterImplementation(selector.SubsampleProverAction{})
	RegisterImplementation(selector.SubsampleVerifierAction{})

	RegisterImplementation(expr_handle.ExprHandleProverAction{})

	RegisterImplementation(plonkinternal.CheckingActivators{})

	RegisterImplementation(cmimc.ExternalHasherBuilder{})
	RegisterImplementation(cmimc.ExternalHasherFactory{})

	RegisterImplementation(plonkinternal.CheckingActivators{})
	RegisterImplementation(plonkinternal.InitialBBSProverAction{})
	RegisterImplementation(plonkinternal.PlonkNoCommitProverAction{})
	RegisterImplementation(plonkinternal.LROCommitProverAction{})

	RegisterImplementation(fr.Element{})

	RegisterImplementation(dedicated.StackedColumn{})

	// Distributed modules
	RegisterImplementation(distributed.AssignLPPQueries{})
	RegisterImplementation(distributed.SetInitialFSHash{})
	RegisterImplementation(distributed.CheckNxHash{})
	RegisterImplementation(distributed.StandardModuleDiscoverer{})
	RegisterImplementation(distributed.LppWitnessAssignment{})
	RegisterImplementation(distributed.ModuleGLAssignGL{})
	RegisterImplementation(distributed.ModuleGLAssignSendReceiveGlobal{})
	RegisterImplementation(distributed.ModuleGLCheckSendReceiveGlobal{})
	RegisterImplementation(distributed.LPPSegmentBoundaryCalculator{})
	RegisterImplementation(distributed.ConglomerationHierarchicalVerifierAction{})
}

var (
	typeToID = make(map[reflect.Type]core.TypeID)
	idToType = make(map[core.TypeID]reflect.Type)
	regLock  sync.RWMutex
)

// Re-use your existing init() logic, but call this instead of collection.NewMapping
func RegisterImplementation(instance any) {
	t := reflect.TypeOf(instance)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	regLock.Lock()
	defer regLock.Unlock()

	// Deterministic ID assignment is hard without a fixed list.
	// For now, simple incremental.
	// IN PRODUCTION: Sort types by name and assign IDs deterministically at startup.
	if _, exists := typeToID[t]; exists {
		return
	}

	id := core.TypeID(len(typeToID) + 1)
	typeToID[t] = id
	idToType[id] = t
}

func getTypeID(t reflect.Type) (core.TypeID, bool) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	regLock.RLock()
	defer regLock.RUnlock()
	id, ok := typeToID[t]
	return id, ok
}

func getTypeByID(id core.TypeID) (reflect.Type, bool) {
	regLock.RLock()
	defer regLock.RUnlock()
	t, ok := idToType[id]
	return t, ok
}
