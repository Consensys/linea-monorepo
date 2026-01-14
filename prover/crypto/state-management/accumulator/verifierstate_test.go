package accumulator_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifierStateHikesItsOwnState(t *testing.T) {

	var (
		proverState   = newTestAccumulatorPoseidon2DummyVal()
		verifierState = proverState.VerifierState()
	)

	preProver, preVerifier := proverState.TopRoot(), verifierState.TopRoot()
	fmt.Printf("prover=%v verifier=%v\n", preProver.Hex(), preVerifier.Hex())

	tr := proverState.InsertAndProve(dumkey(0), dumkey(1))
	if err := verifierState.VerifyInsertion(tr); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, proverState.NextFreeNode, verifierState.NextFreeNode)
	assert.Equal(t, proverState.SubTreeRoot(), verifierState.SubTreeRoot)

	// Check if the tree root is correct
	assert.Equal(t, proverState.TopRoot(), verifierState.TopRoot())

	postProver, postVerifier := proverState.TopRoot(), verifierState.TopRoot()
	fmt.Printf("prover=%v verifier=%v\n", postProver.Hex(), postVerifier.Hex())

}
