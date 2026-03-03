package aggregation

import (
	"fmt"

	frBn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/finalwrap"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/tree"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

// TreeRequest contains the inputs for the tree aggregation pipeline.
// Unlike the standard Request (which carries serialized BLS12-377 PLONK proofs),
// this carries in-memory recursion witnesses from the execution provers.
type TreeRequest struct {
	// LeafWitnesses are the recursion.Witness values extracted from execution
	// provers via ProveForTree. Each witness corresponds to one execution proof.
	LeafWitnesses []recursion.Witness

	// TreeAgg is the compiled tree aggregation structure (compiled at setup time).
	TreeAgg *tree.TreeAggregation

	// PublicInput is the aggregation public input hash in hex string format.
	PublicInput string
}

// ProveTree is the alternative entry point for the tree aggregation pipeline.
// It replaces the standard Prove flow (PI → BW6 → BN254) with:
//  1. Tree aggregation of leaf witnesses (KoalaBear wizard proofs)
//  2. BN254 final wrap with KoalaBear field emulation
//
// The caller is responsible for:
//   - Running execution provers with ProveForTree to get leaf witnesses
//   - Compiling the TreeAggregation at setup time
//   - Computing the aggregation public input hash
func ProveTree(cfg *config.Config, req *TreeRequest) (string, error) {
	return ProveTreeAggregation(cfg, req.TreeAgg, req.LeafWitnesses, req.PublicInput)
}

// ProveTreeAggregation implements the new 3-stage proof pipeline:
//  1. Inner proofs: Vortex on KoalaBear with Ring-SIS + Poseidon2 (already done)
//  2. Tree aggregation: Binary tree of KoalaBear wizard proofs
//  3. Final wrap: BN254 SNARK with KoalaBear field emulation
//
// It receives pre-computed leaf witnesses from execution provers and produces
// a BN254 proof suitable for Ethereum verification.
//
// If treeAgg is nil, it will be compiled on-the-fly from the config.
func ProveTreeAggregation(
	cfg *config.Config,
	treeAgg *tree.TreeAggregation,
	leafWitnesses []recursion.Witness,
	publicInput string,
) (string, error) {

	if cfg.Aggregation.ProverMode == config.ProverModeDev {
		return makeDummyProof(cfg, publicInput, circuits.MockCircuitIDEmulation), nil
	}

	// Compile tree aggregation on-the-fly if not provided
	if treeAgg == nil {
		var err error
		treeAgg, err = compileTreeAggFromConfig(cfg)
		if err != nil {
			return "", fmt.Errorf("could not compile tree aggregation: %w", err)
		}
	}

	logrus.Infof("Starting tree aggregation with %d leaf witnesses, tree depth %d",
		len(leafWitnesses), treeAgg.Depth())

	// Stage 2: Tree aggregation
	rootProof, err := treeAgg.ProveTree(leafWitnesses)
	if err != nil {
		return "", fmt.Errorf("tree aggregation failed: %w", err)
	}

	logrus.Info("Tree aggregation complete, generating BN254 final wrap proof...")

	// Stage 3: BN254 final wrap
	proof, err := makeTreeBn254Proof(cfg, treeAgg, rootProof, publicInput)
	if err != nil {
		return "", fmt.Errorf("BN254 final wrap failed: %w", err)
	}

	return circuits.SerializeProofSolidityBn254(proof), nil
}

// compileTreeAggFromConfig compiles a tree aggregation on-the-fly using the
// config's trace limits and tree aggregation parameters. It obtains the leaf
// CompiledIOP from the full zkEVM compilation (which is cached via sync.Once).
func compileTreeAggFromConfig(cfg *config.Config) (*tree.TreeAggregation, error) {
	limits := cfg.TracesLimits
	logrus.Info("Compiling execution zkEVM for tree aggregation...")
	fullZkEvm := zkevm.FullZkEvm(&limits, cfg)
	leafComp := fullZkEvm.LeafCompiledIOP()

	maxDepth := cfg.TreeAggregation.MaxDepth
	if maxDepth < 1 {
		maxDepth = 1
	}

	logrus.Infof("Compiling tree aggregation with depth=%d", maxDepth)
	treeAgg := tree.CompileTreeAggregation(leafComp, maxDepth)
	logrus.Infof("Tree aggregation compiled: %d levels", treeAgg.Depth())

	return treeAgg, nil
}

// makeTreeBn254Proof generates the BN254 final wrap proof that verifies
// the tree aggregation root wizard proof using KoalaBear field emulation.
func makeTreeBn254Proof(
	cfg *config.Config,
	treeAgg *tree.TreeAggregation,
	rootProof wizard.Proof,
	publicInput string,
) (plonk.Proof, error) {

	logrus.Info("Loading BN254 final wrap setup...")
	setup, err := circuits.LoadSetup(cfg, circuits.FinalWrapCircuitID)
	if err != nil {
		return nil, fmt.Errorf("could not load final wrap setup: %w", err)
	}

	var piBn254 frBn254.Element
	piBytes, err := hexutil.Decode(publicInput)
	if err != nil {
		return nil, fmt.Errorf("could not decode public input: %w", err)
	}
	piBn254.SetBytes(piBytes)

	logrus.Info("Running BN254 final wrap prover...")
	proof, err := finalwrap.MakeProof(
		&setup,
		treeAgg.RootCompiledIOP(),
		rootProof,
		piBn254,
	)
	if err != nil {
		return nil, fmt.Errorf("BN254 final wrap proof failed: %w", err)
	}

	return proof, nil
}
