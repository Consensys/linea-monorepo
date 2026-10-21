package fri

import (
	"math/rand/v2"
	"testing"

	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/crypto/koalabear/poseidon2"
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/maths/koalabear/field"
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/utils"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/stretchr/testify/require"
)

// pairLeavesWitness is groupedWitness with its largest size, 3, split into two
// groups as well, which only pair leaves allow.
func pairLeavesWitness(prng *rand.Rand) Batch {
	witness := groupedWitness(prng)
	witness[3] = SizedTable{
		Base:   [][]field.Element{field.VecPseudoRandBase(prng, 8)},
		Ext:    [][]field.Ext{field.VecPseudoRandExt(prng, 8)},
		Groups: []GroupShape{{ExtWidth: 1}, {BaseWidth: 1}},
	}
	return witness
}

// newPairLeavesPCSFixture commits pairLeavesWitness with pair leaves, opens it
// as numBatches batches sharing one input tree at four queries, and checks the
// native verifier accepts.
func newPairLeavesPCSFixture(t *testing.T, numBatches int) pcsOpenVerifyFixture {
	t.Helper()
	params, err := NewParams(4, 3, 4)
	require.NoError(t, err)
	pcs, err := NewPCS(params, makeEncoders(4, 2))
	require.NoError(t, err)
	pcs.PairLeaves = true
	witness := pairLeavesWitness(rand.New(utils.NewRandSource(20261022)))
	committed := pcs.Commit(witness)
	require.True(t, committed.PairLeaves)

	shifts := BatchShifts{
		{Base: [][]int{{0}, {0}}},
		{Base: [][]int{{0}}, Ext: [][]int{{0}}},
		{Base: [][]int{{0}, {1}}, Ext: [][]int{{0}}},
		{Base: [][]int{{0}}, Ext: [][]int{{1}}},
	}
	challenges := Challenges{
		FoldAlphas: []field.Ext{
			field.UintsToExt(7, 1, 0, 0, 0, 0),
			field.UintsToExt(11, 0, 1, 0, 0, 0),
			field.UintsToExt(13, 0, 0, 1, 0, 0),
		},
		QueryPositions: []int{5, 9, 2, 14},
	}
	zeta := field.UintsToExt(19, 2, 3, 5, 7, 11)
	repeat := make([]int, numBatches)
	proof, claimed := openForTest(t, pcs, openInputs{
		Witnesses:  utils.Map(func(int) Batch { return witness }, repeat),
		Committed:  utils.Map(func(int) CommitterState { return committed }, repeat),
		Shifts:     utils.Map(func(int) BatchShifts { return shifts }, repeat),
		Zeta:       zeta,
		Challenges: challenges,
	})
	shape := witness.Shape()
	fx := pcsOpenVerifyFixture{
		pcs:       pcs,
		committed: []CommitterState{committed},
		input: VerifyInputs{
			Roots:         utils.Map(func(int) field.Octuplet { return committed.Tree.Root() }, repeat),
			Shapes:        utils.Map(func(int) Shape { return shape }, repeat),
			Shifts:        utils.Map(func(int) BatchShifts { return shifts }, repeat),
			ClaimedValues: claimed,
			Zeta:          zeta,
			Challenges:    challenges,
		},
		proof: proof,
	}
	require.NoError(t, pcs.Verify(fx.input, fx.proof), "native verifier must accept the pair-leaf fixture")
	return fx
}

// TestPairLeavesTree checks the pair-leaf tree against its definition: the
// largest size is digested as row pairs, group by group, and leaf j is the
// first group's digest with the others chained after it; every other size is
// injected as before. The tree is one level shorter than with row leaves.
func TestPairLeavesTree(t *testing.T) {
	witness := pairLeavesWitness(rand.New(utils.NewRandSource(8)))
	encoded := witness.Encode(makeEncoders(4, 2))
	tree := encoded.merkleize(true)

	groupsOf := func(table SizedTable) []SizedTable {
		groups := sizeGroups(len(table.Base), len(table.Ext), table.Groups)
		res := make([]SizedTable, 0, len(groups))
		var base, ext int
		for _, g := range groups {
			res = append(res, SizedTable{Base: table.Base[base : base+g.BaseWidth], Ext: table.Ext[ext : ext+g.ExtWidth]})
			base += g.BaseWidth
			ext += g.ExtWidth
		}
		return res
	}

	bottom := encoded[len(encoded)-1]
	leaves := make([]field.Octuplet, bottom.Size()/2)
	for j := range leaves {
		groups := groupsOf(bottom)
		leaves[j] = referenceGroupDigest(groups[0], j)
		for _, group := range groups[1:] {
			leaves[j] = poseidon2.Compress(leaves[j], referenceGroupDigest(group, j))
		}
	}
	injections := make([][][]field.Octuplet, utils.Log2Ceil(len(leaves)))
	for i := range len(encoded) - 1 {
		level := utils.Log2Ceil(encoded[i].Size() / 2)
		for _, group := range groupsOf(encoded[i]) {
			digests := make([]field.Octuplet, encoded[i].Size()/2)
			for j := range digests {
				digests[j] = referenceGroupDigest(group, j)
			}
			injections[level] = append(injections[level], digests)
		}
	}
	want := allocTree(len(leaves))
	copy(want.Nodes[len(leaves)-1:], leaves)
	want.buildLevels(injections)
	require.Equal(t, want.Root(), tree.Root(), "pair-leaf root must follow its definition")
	require.Equal(t, bottom.Size()/2, tree.NumLeaves(), "a leaf digests a row pair")

	// A single-group largest size: the leaf is that group's digest, unchained.
	single := make(Batch, 3)
	single[2] = SizedTable{Ext: [][]field.Ext{field.VecPseudoRandExt(rand.New(utils.NewRandSource(9)), 4)}}
	enc := single.Encode(makeEncoders(3, 2))
	singleTree := enc.merkleize(true)
	for j := range singleTree.NumLeaves() {
		require.Equal(t, referenceGroupDigest(enc[2], j), singleTree.Nodes[singleTree.NumLeaves()-1+j])
	}

	// The smallest largest size, two encoded rows, makes a one-leaf tree.
	tiny := Batch{{Base: [][]field.Element{field.VecPseudoRandBase(rand.New(utils.NewRandSource(10)), 1)}}}
	tinyEnc := tiny.Encode(makeEncoders(1, 2))
	tinyTree := tinyEnc.merkleize(true)
	require.Equal(t, 1, tinyTree.NumLevel())
	require.Equal(t, referenceGroupDigest(tinyEnc[0], 0), tinyTree.Root())
}

// TestPCSVerifyPairLeaves opens a pair-leaf commitment whose largest size holds
// two groups, and checks that the verifier rejects it under row leaves, without
// the largest size's groups, or with a malformed bottom pair.
func TestPCSVerifyPairLeaves(t *testing.T) {
	fx := newPairLeavesPCSFixture(t, 2)

	info, err := inputCapShapeInfo(fx.pcs.Params, fx.input.Shapes[0], true)
	require.NoError(t, err)
	require.Equal(t, info.height-1, info.treeHeight())
	require.Len(t, info.levelGroups[info.height-1], 2, "the largest size must hold several groups")
	_, err = inputCapShapeInfo(fx.pcs.Params, fx.input.Shapes[0], false)
	require.Error(t, err, "row leaves must refuse a grouped largest size")

	rowLeaves := *fx.pcs
	rowLeaves.PairLeaves = false
	require.Error(t, rowLeaves.Verify(fx.input, fx.proof), "row-leaf verification must reject a pair-leaf proof")

	flatBottom := append(Shape(nil), fx.input.Shapes[0]...)
	flatBottom[3].Groups = nil
	in := fx.input
	in.Shapes = []Shape{flatBottom, flatBottom}
	require.Error(t, fx.pcs.Verify(in, fx.proof), "the largest size's grouping must bind")

	malformed := fx.proof
	malformed.InputQueries = append([]InputQuery(nil), fx.proof.InputQueries...)
	malformed.InputQueries[0] = append(InputQuery(nil), fx.proof.InputQueries[0]...)
	opening := malformed.InputQueries[0][0]
	opening.Leaves = append([]*RowPair(nil), opening.Leaves...)
	pair := *opening.Leaves[len(opening.Leaves)-1]
	pair[1].Base = nil
	opening.Leaves[len(opening.Leaves)-1] = &pair
	malformed.InputQueries[0][0] = opening
	require.NotPanics(t, func() {
		require.Error(t, fx.pcs.Verify(fx.input, malformed), "a malformed bottom pair must be rejected")
	})
}

// TestPCSAddOpeningRejectsLeafModeMismatch checks that a commitment can only be
// opened by a PCS of the same leaf mode.
func TestPCSAddOpeningRejectsLeafModeMismatch(t *testing.T) {
	fx := newPairLeavesPCSFixture(t, 1)
	rowLeaves, err := NewPCS(fx.pcs.Params, fx.pcs.Encoders)
	require.NoError(t, err)
	err = rowLeaves.AddOpening(fx.committed[0], fx.input.Zeta, fx.input.Shifts[0], fx.input.ClaimedValues[0])
	require.ErrorContains(t, err, "pair leaves")
}

// TestPCSVerifyGnarkPairLeaves is the circuit counterpart: the circuit accepts
// the pair-leaf proof, rejects a tampered bottom row, and cannot even be built
// under row leaves, which refuse a grouped largest size.
func TestPCSVerifyGnarkPairLeaves(t *testing.T) {
	fx := newPairLeavesPCSFixture(t, 1)
	template, assignment := circuitFor(fx, fx.proof)
	require.NoError(t, solvePCSCircuit(t, template, assignment), "pair-leaf proof must satisfy the verifier circuit")

	tampered := fx.proof
	tampered.InputQueries = append([]InputQuery(nil), fx.proof.InputQueries...)
	tampered.InputQueries[0] = append(InputQuery(nil), fx.proof.InputQueries[0]...)
	opening := tampered.InputQueries[0][0]
	opening.Leaves = append([]*RowPair(nil), opening.Leaves...)
	pair := *opening.Leaves[len(opening.Leaves)-1]
	pair[1].Ext = append([]field.Ext(nil), pair[1].Ext...)
	pair[1].Ext[0] = field.UintsToExt(1, 2, 3, 4, 5, 6)
	opening.Leaves[len(opening.Leaves)-1] = &pair
	tampered.InputQueries[0][0] = opening
	template, assignment = circuitFor(fx, tampered)
	require.Error(t, solvePCSCircuit(t, template, assignment), "a tampered bottom row must not satisfy the circuit")

	rowLeaves := *fx.pcs
	rowLeaves.PairLeaves = false
	fx.pcs = &rowLeaves
	template, _ = circuitFor(fx, fx.proof)
	_, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, template)
	require.ErrorContains(t, err, "the largest size must hold a single group")
}
