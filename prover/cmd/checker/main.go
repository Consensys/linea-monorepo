package main

import (
	"flag"

	"github.com/consensys/accelerated-crypto-monorepo/backend"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	traceFile := flag.String("trace", "", "the trace file to check (in json.gz format)")
	flag.Parse()

	if *traceFile == "" {
		utils.Panic("missing --trace argument")
		return
	}

	logrus.Infof("Checking %v", *traceFile)
	backend.RunCorsetAndChecker(*traceFile)
}
