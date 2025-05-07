package main

import (
	"context"

	"github.com/consensys/linea-monorepo/prover/cmd/prover/cmd"
)

func main() {
	setup()
	//prove()
}

const configFile = "/home/ubuntu/linea-monorepo/prover/integration/doubledict/config-sepolia-full-doubledict.toml"

func setup() {
	assertNoError(cmd.Setup(context.TODO(), cmd.SetupArgs{
		Circuits:   "blob-decompression-v1",
		DictPath:   "",
		DictSize:   65536,
		AssetsDir:  "/home/ubuntu/linea-monorepo/prover/prover-assets",
		ConfigFile: configFile,
	}))
}

func prove() {
	assertNoError(cmd.Prove(cmd.ProverArgs{
		Input:      "/home/ubuntu/linea-monorepo/prover/integration/doubledict/11383347-11384215-bcv0.0-ccv0.0-7f7b2c7fcdd136111a6acdaf9c34c278dc56e306365b8141d55e0f8fd9f418cb-getZkBlobCompressionProof.json",
		Output:     "",
		ConfigFile: configFile,
	}))
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}
