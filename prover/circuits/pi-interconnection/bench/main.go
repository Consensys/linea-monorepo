package main

import (
	"fmt"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/consensys/linea-monorepo/prover/circuits/internal/test_utils"
	pi_interconnection "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection"
	pitesting "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/test_utils"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"

	//"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/test_utils"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/stretchr/testify/assert"
)

func main() {
	var b test_utils.FakeTestingT
	req := pitesting.AssignSingleBlockBlob(b)

	c, err := pi_interconnection.Config{
		MaxNbDecompression:   400,
		MaxNbExecution:       400,
		MaxNbKeccakF:         10000,
		MaxNbMsgPerExecution: 16,
		L2MsgMerkleDepth:     5,
		L2MessageMaxNbMerkle: 10,
	}.Compile(dummy.Compile) // note that the solving/proving time will not reflect the wizard proof or verification
	assert.NoError(b, err)

	a, err := c.Assign(req)
	assert.NoError(b, err)

	c.Circuit.UseGkrMimc = true

	cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, c.Circuit, frontend.WithCapacity(40_000_000))
	assert.NoError(b, err)

	kzgc, kzgl, err := unsafekzg.NewSRS(cs)
	assert.NoError(b, err)

	pk, _, err := plonk.Setup(cs, kzgc, kzgl)
	assert.NoError(b, err)

	secondsStart := time.Now().Unix()

	w, err := frontend.NewWitness(&a, ecc.BLS12_377.ScalarField())
	assert.NoError(b, err)
	_, err = plonk.Prove(cs, pk, w, backend.WithSolverOptions(gkrmimc.SolverOpts(cs)...))
	assert.NoError(b, err)

	fmt.Println(time.Now().Unix()-secondsStart, "seconds")
}
