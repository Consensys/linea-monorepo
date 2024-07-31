// The bw6circuit package provides an implementation of the BW6 proof aggregation
// circuit. This circuits can aggregate several PLONK proofs.
package aggregation

import (
	"fmt"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/sw_bls12377"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/rangecheck"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection/keccak"
	public_input "github.com/consensys/zkevm-monorepo/prover/public-input"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
	"slices"
)

// shorthand for the emulated types as this can get verbose very quickly with
// generics. `em` stands for emulated
type (
	emFr       = sw_bls12377.ScalarField
	emG1       = sw_bls12377.G1Affine
	emG2       = sw_bls12377.G2Affine
	emGT       = sw_bls12377.GT
	emProof    = emPlonk.Proof[emFr, emG1, emG2]
	emVkey     = emPlonk.VerifyingKey[emFr, emG1, emG2]
	emCircVKey = emPlonk.CircuitVerifyingKey[emFr, emG1]
	emWitness  = emPlonk.Witness[emFr]
	// emBaseVkey = emPlonk.BaseVerifyingKey[emFr, emG1, emG2]
)

// The AggregationCircuit is used to aggregate multiple execution proofs and
// aggregation proofs together.
type AggregationCircuit struct {

	// The list of claims to be provided to the circuit.
	ProofClaims []proofClaim `gnark:",secret"`
	// List of available verifying keys that are available to the circuit. This
	// is treated as a constant by the circuit.
	verifyingKeys []emVkey `gnark:"-"`

	// Dummy general public input
	DummyPublicInput frontend.Variable `gnark:",public"`
}

func (c *AggregationCircuit) Define(api frontend.API) error {
	// Verify the constraints the execution proofs
	err := verifyClaimBatch(api, c.verifyingKeys, c.ProofClaims)
	if err != nil {
		return fmt.Errorf("processing execution proofs: %w", err)
	}

	// TODO incorporate statements (circuitID+publicInput?) and vk's into public input

	return err
}

// Instantiate a new AggregationCircuit from a list of verification keys and
// a maximal number of proofs. The function should only be called with the
// purpose of running `frontend.Compile` over it.
func AllocateAggregationCircuit(
	nbProofs int,
	verifyingKeys []plonk.VerifyingKey,
) (*AggregationCircuit, error) {

	var (
		err           error
		emVKeys       = make([]emVkey, len(verifyingKeys))
		csPlaceHolder = getPlaceHolderCS()
		proofClaims   = make([]proofClaim, nbProofs)
	)

	for i := range verifyingKeys {
		emVKeys[i], err = emPlonk.ValueOfVerifyingKey[emFr, emG1, emG2](verifyingKeys[i])
		if err != nil {
			return nil, fmt.Errorf("while converting the verifying key #%v (execution) into its emulated gnark version: %w", i, err)
		}
	}

	for i := range proofClaims {
		proofClaims[i] = allocatableClaimPlaceHolder(csPlaceHolder)
	}

	return &AggregationCircuit{
		verifyingKeys: emVKeys,
		ProofClaims:   proofClaims,
	}, nil

}

func verifyClaimBatch(api frontend.API, vks []emVkey, claims []proofClaim) error {
	verifier, err := emPlonk.NewVerifier[emFr, emG1, emG2, emGT](api)
	if err != nil {
		return fmt.Errorf("while instantiating the verifier: %w", err)
	}

	var (
		bvk       = vks[0].BaseVerifyingKey
		cvks      = make([]emCircVKey, len(vks))
		switches  = make([]frontend.Variable, len(claims))
		proofs    = make([]emProof, len(claims))
		witnesses = make([]emWitness, len(claims))
	)

	for i := range vks {
		cvks[i] = vks[i].CircuitVerifyingKey
	}

	for i := range claims {
		proofs[i] = claims[i].Proof
		switches[i] = claims[i].CircuitID
		witnesses[i] = claims[i].PublicInput
	}

	err = verifier.AssertDifferentProofs(bvk, cvks, switches, proofs, witnesses, emPlonk.WithCompleteArithmetic())
	if err != nil {
		return fmt.Errorf("AssertDifferentProofs returned an error: %w", err)
	}
	return nil
}

type FunctionalPublicInputQSnark struct {
	ParentShnarf             [32]frontend.Variable
	NbDecompression          frontend.Variable
	InitialStateRootHash     frontend.Variable
	InitialBlockNumber       frontend.Variable
	InitialBlockTimestamp    frontend.Variable
	InitialRollingHash       [32]frontend.Variable
	InitialRollingHashNumber frontend.Variable
	ChainID                  frontend.Variable // for now we're forcing all executions to have the same chain ID
	L2MessageServiceAddr     frontend.Variable // 20 bytes
}

type FunctionalPublicInputSnark struct {
	FunctionalPublicInputQSnark
	NbL2Messages         frontend.Variable // TODO not used in hash. delete if not necessary
	L2MsgMerkleTreeRoots [][32]frontend.Variable
	// FinalStateRootHash     frontend.Variable redundant: incorporated into final shnarf
	FinalBlockNumber       frontend.Variable
	FinalBlockTimestamp    frontend.Variable
	FinalRollingHash       [32]frontend.Variable
	FinalRollingHashNumber frontend.Variable
	FinalShnarf            [32]frontend.Variable
	L2MsgMerkleTreeDepth   int
}

// FunctionalPublicInput holds the same info as public_input.Aggregation, except in parsed form
type FunctionalPublicInput struct {
	ParentShnarf             [32]byte
	NbDecompression          uint64
	InitialStateRootHash     [32]byte
	InitialBlockNumber       uint64
	InitialBlockTimestamp    uint64
	InitialRollingHash       [32]byte
	InitialRollingHashNumber uint64
	ChainID                  uint64 // for now we're forcing all executions to have the same chain ID
	L2MessageServiceAddr     types.EthAddress
	NbL2Messages             uint64 // TODO not used in hash. delete if not necessary
	L2MsgMerkleTreeRoots     [][32]byte
	//FinalStateRootHash       [32]byte		redundant: incorporated into shnarf
	FinalBlockNumber       uint64
	FinalBlockTimestamp    uint64
	FinalRollingHash       [32]byte
	FinalRollingHashNumber uint64
	FinalShnarf            [32]byte
	L2MsgMerkleTreeDepth   int
}

// NewFunctionalPublicInput does NOT set all fields, only the ones covered in public_input.Aggregation
func NewFunctionalPublicInput(fpi *public_input.Aggregation) (s *FunctionalPublicInput, err error) {
	s = &FunctionalPublicInput{
		InitialBlockNumber:       uint64(fpi.LastFinalizedBlockNumber),
		InitialBlockTimestamp:    uint64(fpi.ParentAggregationLastBlockTimestamp),
		InitialRollingHashNumber: uint64(fpi.LastFinalizedL1RollingHashMessageNumber),
		L2MsgMerkleTreeRoots:     make([][32]byte, len(fpi.L2MsgRootHashes)),
		FinalBlockNumber:         uint64(fpi.FinalBlockNumber),
		FinalBlockTimestamp:      uint64(fpi.FinalTimestamp),
		FinalRollingHashNumber:   uint64(fpi.L1RollingHashMessageNumber),
		L2MsgMerkleTreeDepth:     fpi.L2MsgMerkleTreeDepth,
	}

	if err = copyFromHex(s.InitialStateRootHash[:], fpi.ParentStateRootHash); err != nil {
		return
	}
	if err = copyFromHex(s.FinalRollingHash[:], fpi.L1RollingHash); err != nil {
		return
	}
	if err = copyFromHex(s.InitialRollingHash[:], fpi.LastFinalizedL1RollingHash); err != nil {
		return
	}
	if err = copyFromHex(s.ParentShnarf[:], fpi.ParentAggregationFinalShnarf); err != nil {
		return
	}
	if err = copyFromHex(s.FinalShnarf[:], fpi.FinalShnarf); err != nil {
		return
	}

	for i := range s.L2MsgMerkleTreeRoots {
		if err = copyFromHex(s.L2MsgMerkleTreeRoots[i][:], fpi.L2MsgRootHashes[i]); err != nil {
			return
		}
	}
	return
}

func (pi *FunctionalPublicInput) ToSnarkType() FunctionalPublicInputSnark {
	s := FunctionalPublicInputSnark{
		FunctionalPublicInputQSnark: FunctionalPublicInputQSnark{
			InitialBlockNumber:       pi.InitialBlockNumber,
			InitialBlockTimestamp:    pi.InitialBlockTimestamp,
			InitialRollingHash:       [32]frontend.Variable{},
			InitialRollingHashNumber: pi.InitialRollingHashNumber,
			InitialStateRootHash:     pi.InitialStateRootHash[:],

			NbDecompression:      pi.NbDecompression,
			ChainID:              pi.ChainID,
			L2MessageServiceAddr: pi.L2MessageServiceAddr[:],
		},
		L2MsgMerkleTreeRoots:   make([][32]frontend.Variable, len(pi.L2MsgMerkleTreeRoots)),
		FinalBlockNumber:       pi.FinalBlockNumber,
		FinalBlockTimestamp:    pi.FinalBlockTimestamp,
		FinalRollingHashNumber: pi.FinalRollingHashNumber,
		L2MsgMerkleTreeDepth:   pi.L2MsgMerkleTreeDepth,
	}

	internal.Copy(s.FinalRollingHash[:], pi.FinalRollingHash[:])
	internal.Copy(s.InitialRollingHash[:], pi.InitialRollingHash[:])
	internal.Copy(s.ParentShnarf[:], pi.ParentShnarf[:])
	internal.Copy(s.FinalShnarf[:], pi.FinalShnarf[:])

	for i := range s.L2MsgMerkleTreeRoots {
		internal.Copy(s.L2MsgMerkleTreeRoots[i][:], pi.L2MsgMerkleTreeRoots[i][:])
	}

	return s
}

func (pi *FunctionalPublicInputSnark) Sum(api frontend.API, hash keccak.BlockHasher) [32]frontend.Variable {
	// number of hashes: 12
	sum := hash.Sum(nil,
		pi.ParentShnarf,
		pi.FinalShnarf,
		internal.ToBytes(api, pi.InitialBlockTimestamp),
		internal.ToBytes(api, pi.FinalBlockTimestamp),
		internal.ToBytes(api, pi.InitialBlockNumber),
		internal.ToBytes(api, pi.FinalBlockNumber),
		pi.InitialRollingHash,
		pi.FinalRollingHash,
		internal.ToBytes(api, pi.InitialRollingHashNumber),
		internal.ToBytes(api, pi.FinalRollingHashNumber),
		internal.ToBytes(api, pi.L2MsgMerkleTreeDepth),
		hash.Sum(nil, pi.L2MsgMerkleTreeRoots...),
	)

	// turn the hash into a bn254 element
	var res [32]frontend.Variable
	copy(res[:], internal.ReduceBytes[emulated.BN254Fr](api, sum[:]))
	return res
}

func (pi *FunctionalPublicInputQSnark) RangeCheck(api frontend.API) {
	rc := rangecheck.New(api)
	for _, v := range append(slices.Clone(pi.InitialRollingHash[:]), pi.ParentShnarf[:]...) {
		rc.Check(v, 8)
	}
	// not checking L2MsgServiceAddr as its range is never assumed in the pi circuit
	// not checking NbDecompressions as the NewRange in the pi circuit range checks it; TODO do it here instead
}

func copyFromHex(dst []byte, src string) error {
	b, err := utils.HexDecodeString(src)
	if err != nil {
		return err
	}
	copy(dst[len(dst)-len(b):], b) // panics if src is too long
	return nil
}
