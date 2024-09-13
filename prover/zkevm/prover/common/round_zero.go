package common

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// It commits to the columns of round zero, from the same module.
func CreateColFn(comp *wizard.CompiledIOP, rootName string, size int) func(name string, args ...interface{}) ifaces.Column {

	return func(name string, args ...interface{}) ifaces.Column {
		s := []string{rootName, name}

		v := strings.Join(s, "_")

		return comp.InsertCommit(0, ifaces.ColIDf(v, args...), size)
	}

}
