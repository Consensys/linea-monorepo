package main

import (
	"flag"

	"github.com/consensys/accelerated-crypto-monorepo/backend"
	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

func main() {
	inFile := flag.String("in", "", "input trace filepath")
	outFile := flag.String("out", "", "output proof filepath")
	flag.Parse()

	if *inFile == "" {
		utils.Panic("missing input file")
		return
	}
	if *outFile == "" {
		utils.Panic("missing output file")
		return
	}

	config.InitApp()
	backend.RunCorsetAndProver(*inFile, *outFile)
}
