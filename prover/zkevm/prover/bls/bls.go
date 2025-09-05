package bls

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type group int

const (
	G1 group = iota
	G2
)

type membership int

const (
	CURVE membership = iota
	GROUP
)

const ROUND_NR = 0

func (g group) String() string {
	switch g {
	case G1:
		return "G1"
	case G2:
		return "G2"
	default:
		panic("unknown group")
	}
}

func (g group) StringCurve() string {
	switch g {
	case G1:
		return "C1"
	case G2:
		return "C2"
	default:
		panic("unknown group")
	}
}

func (g group) StringMap() string {
	switch g {
	case G1:
		return "FP"
	case G2:
		return "FP2"
	default:
		panic("unknown group")
	}
}

func createColFn(comp *wizard.CompiledIOP, rootName string, size int) func(name string) ifaces.Column {
	return func(name string) ifaces.Column {
		return comp.InsertCommit(ROUND_NR, ifaces.ColIDf("%s_%s", rootName, name), size)
	}
}
