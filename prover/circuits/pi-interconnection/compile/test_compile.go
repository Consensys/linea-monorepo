package main

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal/test_utils"
	pi_interconnection "github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/stretchr/testify/assert"
)

func main() {

	fmt.Println("creating wizard circuit")

	c, err := pi_interconnection.Config{
		MaxNbDecompression:   400,
		MaxNbExecution:       400,
		MaxNbKeccakF:         10000,
		MaxNbMsgPerExecution: 16,
		L2MsgMerkleDepth:     5,
		L2MessageMaxNbMerkle: 10,
	}.Compile(dummy.Compile)

	var t test_utils.FakeTestingT
	assert.NoError(t, err)

	p := profile.Start()
	_, err = frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, c.Circuit, frontend.WithCapacity(1<<27))
	p.Stop()
	assert.NoError(t, err)
}
