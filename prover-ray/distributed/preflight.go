package distributed

import (
	"github.com/consensys/linea-monorepo/prover-ray/arithmetization"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
)

// LPPCommitment pairs a segment identity with the Poseidon2 Merkle root of
// its LPP column data. It is the prover's native computation of what the GL
// circuit will later expose as a public input (lppMerkleRootPublicInput).
//
// By computing this from preflight column data — before any GL proof runs —
// the prover can derive shared randomness as soon as arithmetization signals
// preflight completion, eliminating the serial GL → shared-randomness → LPP
// dependency.
type LPPCommitment struct {
	ModuleIndex  int
	SegmentIndex int
	// Hash is the Poseidon2 Merkle root over the LPP columns of this segment.
	// It must be computed by the same scheme the GL circuit uses so that the
	// GL proof's public output is consistent with what was used for shared
	// randomness.
	Hash field.Octuplet
}

// CommitLPPColumns computes the LPP commitment for a single preflight segment.
//
// It hashes the segment's LPP column data using the same Poseidon2 Merkle tree
// scheme that the GL circuit uses internally. No proof is produced here; this
// is a native (non-circuit) computation that the prover runs to get the
// commitment value early.
//
// PSEUDO: the actual hash call would be poseidon2.MerkleRoot(columns...) or
// equivalent; replace the panic with the real Poseidon2 commitment once the
// native hash function is exposed from the crypto package.
func CommitLPPColumns(seg arithmetization.PreflightSegment) LPPCommitment {
	// PSEUDO: concatenate column data in a deterministic canonical order
	// (sorted by ColumnID) and compute the Poseidon2 Merkle root.
	//
	//   sortedIDs := sortedKeys(seg.Columns)
	//   rows := []field.Vector{}
	//   for _, id := range sortedIDs {
	//       rows = append(rows, seg.Columns[id])
	//   }
	//   hash := poseidon2.MerkleRoot(rows...)
	//
	hash := pseudoPoseidon2MerkleRoot(seg.Columns)
	return LPPCommitment{
		ModuleIndex:  seg.ModuleIndex,
		SegmentIndex: seg.SegmentIndex,
		Hash:         hash,
	}
}

// GetSharedRandomness derives the shared Fiat-Shamir randomness used by all
// LPP provers from the set of LPP commitments produced by CommitLPPColumns.
//
// The derivation is order-independent: commitments are inserted into a
// multiset hash together with their (moduleIndex, segmentIndex) so that the
// result does not depend on the order in which they are received (e.g., from
// concurrent goroutines).
//
// The result is placed in ModuleWitnessLPP.InitialFiatShamirState for every
// LPP segment before the LPP prover runs.
func GetSharedRandomness(commitments []LPPCommitment) field.Octuplet {
	// PSEUDO: insert each (moduleIndex, segmentIndex, hash[0..7]) tuple into a
	// multiset hash, then Poseidon2-hash the multiset accumulator.
	//
	//   mset := multsethashing.MSetHash{}
	//   for _, c := range commitments {
	//       mset.Insert(
	//           field.NewElement(uint64(c.ModuleIndex)),
	//           field.NewElement(uint64(c.SegmentIndex)),
	//           c.Hash[0], c.Hash[1], ..., c.Hash[7],
	//       )
	//   }
	//   return poseidon2.HashVec(mset[:]...)
	//
	return pseudoMultisetHashThenPoseidon2(commitments)
}

// CommitAll is a convenience helper that applies CommitLPPColumns to every
// preflight segment and returns the full slice of LPP commitments.
// The caller can then pass the result directly to GetSharedRandomness.
func CommitAll(segs []arithmetization.PreflightSegment) []LPPCommitment {
	commitments := make([]LPPCommitment, len(segs))
	for i, seg := range segs {
		commitments[i] = CommitLPPColumns(seg)
	}
	return commitments
}

// ---------------------------------------------------------------------------
// Pseudo stubs
// ---------------------------------------------------------------------------

func pseudoPoseidon2MerkleRoot(_ map[arithmetization.ColumnID]arithmetization.ColumnData) field.Octuplet {
	panic("pseudo: replace with real Poseidon2 Merkle root over LPP columns")
}

func pseudoMultisetHashThenPoseidon2(_ []LPPCommitment) field.Octuplet {
	panic("pseudo: replace with real multiset hash + Poseidon2")
}
