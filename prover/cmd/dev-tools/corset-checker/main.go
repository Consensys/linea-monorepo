package main

import (
	"fmt"
	"os"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func main() {

	cfg, traceFile, pErr := getParamsFromCLI()
	if pErr != nil {
		fmt.Printf("FATAL\n")
		fmt.Printf("err = %v\n", pErr)
		os.Exit(1)
	}

	var (
		comp  = wizard.Compile(MakeDefine(cfg), dummy.Compile)
		proof = wizard.Prove(comp, MakeProver(traceFile))
		vErr  = wizard.Verify(comp, proof)
	)

	if vErr == nil {
		fmt.Printf("PASSED\n")
	}

	if vErr != nil {
		fmt.Printf("FAILED\n")
		fmt.Printf("err = %v\n", vErr.Error())
		os.Exit(1)
	}

}
