package main

import (
	"github.com/consensys/zkevm-monorepo/prover/cmd/dev-tools/state-manager-inspector/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("exiting with error: %v", err)
	}
}
