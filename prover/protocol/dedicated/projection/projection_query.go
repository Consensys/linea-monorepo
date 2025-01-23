package projection

/*
Package projection implements the utilities for the projection query.

A projection query between sets (columnsA,filterA) and (columnsB,filterB) asserts
whether the columnsA filtered by filterA is the same as columnsB filtered by
filterB, preserving the order.

Example:

FilterA = (1,0,0,1,1), ColumnA := (aO,a1,a2,a3,a4)

FiletrB := (0,0,1,0,0,0,0,0,1,1), ColumnB :=(b0,b1,b2,b3,b4,b5,b6,b7,b8,b9)

Thus we have,

ColumnA filtered by FilterA = (a0,a3,a4)

ColumnB filtered by FilterB  = (b2,b8,b9)

The projection query checks if a0 = b2, a3 = b8, a4 = b9

Note that the query imposes that:
  - the number of 1 in the filters are equal
  - the order of filtered elements is preserved
*/

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
)

func RegisterProjection(
	comp *wizard.CompiledIOP,
	queryName ifaces.QueryID,
	columnsA, columnsB []ifaces.Column,
	filterA, filterB ifaces.Column,
) {

	var (
		round = max(
			wizardutils.MaxRound(columnsA...),
			wizardutils.MaxRound(columnsB...),
			filterA.Round(),
			filterB.Round(),
		)
	)
	comp.InsertProjection(round, queryName, query.ProjectionInput{ColumnA: columnsA, ColumnB: columnsB, FilterA: filterA, FilterB: filterB})
}
