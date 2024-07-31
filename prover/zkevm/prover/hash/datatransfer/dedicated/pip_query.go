package dedicated

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// InsertPartitionedIP registers a partitioned inner-product (PIP) query.
/*
A PIP query is an inner-product query of two  vectors over a given partition.


Example:

partition := (0,..,1,0,...,1,...,0,..1),

where it is 1 on positions k+1,l+1,n.
Note that a partition ends with 1.


colA := (a0,..,ak,b(k+1),...,bl,c(l+1),..cn)

colB := (a'0,..,a'k,b'(k+1),...,b'l,c'(l+1),..c'n)


Then ipTracker is computed as follows;
 -  ipTracker[0] = colA[0]*colB[0]
 -  ipTracker[i]= colA[i]*colB[i] + ipTracker[i-1]*(1-partition[i])

 Where i stands for the row-index.


 The result of PIP is stored in i-th row of ipTracker where partition[i+1] == 1.

 Thus at position (k,l,n-1) we respectively have (\sum_i a_i*a'_i,\sum_i b_i*b'_i,\sum_i c_i*c'_i)
*/
func InsertPartitionedIP(
	comp *wizard.CompiledIOP,
	round int,
	colA, colB ifaces.Column,
	partition, ipTracker ifaces.Column,
) {
	one := symbolic.NewConstant(1)

	// Compute the partitioned inner-product
	cola := ifaces.ColumnAsVariable(colA)
	colb := ifaces.ColumnAsVariable(colB)
	shiftTracker := ifaces.ColumnAsVariable(column.Shift(ipTracker, -1))

	// iptaker[i] = (colA[i] * colB) + ipTracker[i-1]* (1-partition[i]).
	expr1 := ifaces.ColumnAsVariable(ipTracker).
		Sub((cola.Mul(colb)).
			Add(shiftTracker.Mul(one.Sub(ifaces.ColumnAsVariable(partition)))),
		)
	comp.InsertGlobal(round, ifaces.QueryIDf("PIP_%v", ipTracker.GetColID()), expr1)

	// ipTracker[0] =colA[0]*colB[0]
	comp.InsertLocal(round,
		ifaces.QueryIDf("PIP_Local_%v", ipTracker.GetColID()),
		ifaces.ColumnAsVariable(ipTracker).
			Sub(cola.Mul(colb)),
	)
	comp.SubProvers.AppendToInner(round,
		func(run *wizard.ProverRuntime) {
			assignPIP(run, colA, colB, partition, ipTracker)
		},
	)
}

// It assigns IPTracker for PIP.
func assignPIP(run *wizard.ProverRuntime, colA, colB, partition, ipTracker ifaces.Column) {
	cola := colA.GetColAssignment(run).IntoRegVecSaveAlloc()
	colb := colB.GetColAssignment(run).IntoRegVecSaveAlloc()
	partitionWit := partition.GetColAssignment(run).IntoRegVecSaveAlloc()
	one := field.One()
	var notPartition field.Element
	witSize := smartvectors.Density(run.GetColumn(colA.GetColID()))
	var u, v field.Element
	ipTrackerWit := make([]field.Element, witSize)
	if witSize != 0 {
		ipTrackerWit[0].Mul(&cola[0], &colb[0])

		for i := 1; i < witSize; i++ {
			u.Mul(&cola[i], &colb[i])
			notPartition.Sub(&one, &partitionWit[i])
			v.Mul(&ipTrackerWit[i-1], &notPartition)
			ipTrackerWit[i].Add(&u, &v)
		}
	}
	run.AssignColumn(ipTracker.GetColID(), smartvectors.RightZeroPadded(ipTrackerWit, colA.Size()))

}
