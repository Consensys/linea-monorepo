package dedicated

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// InsertPartitionedIP registers a partitioned inner-product (PIP) query.
/*
A PIP query is an inner-product query of two  vectors over a given partition.


Example:

partition := (1,0,0,0,1,0,1,0,0),

colA := (a0,a1,a2,a3,b0,b1,c0,c1,c2)

colB := (a0',a1',a2',a3',b0',b1',c0',c1',c2')


Then ipTracker is computed as follows;
 -  ipTracker[last] = colA[last]*colB[last]
 -  ipTracker[i]= colA[i]*colB[i] + ipTracker[i+1]*(1-partition[i+1])

 Where i stands for the row-index.


 The result of PIP is stored in i-th row of ipTracker where partition[i] == 1.

 Thus at position (k,l,n-1) we respectively have (\sum_i a_i*a'_i,\sum_i b_i*b'_i,\sum_i c_i*c'_i)
*/
func InsertPartitionedIP(
	comp *wizard.CompiledIOP,
	name string,
	colA, colB ifaces.Column,
	partition ifaces.Column,
) ifaces.Column {

	ipTracker := comp.InsertCommit(0, ifaces.ColIDf("%v_%v", name, "IPTracker"), colA.Size())

	// Compute the partitioned inner-product
	// iptaker[i] = (colA[i] * colB[i]) + ipTracker[i+1]* (1-partition[i+1]).
	comp.InsertGlobal(0, ifaces.QueryIDf("PIP_%v", ipTracker.GetColID()),
		sym.Sub(ipTracker,
			sym.Add(
				sym.Mul(colA, colB),
				sym.Mul(
					column.Shift(ipTracker, 1),
					sym.Sub(1, column.Shift(partition, 1)),
				),
			),
		),
	)

	// ipTracker[last] =colA[last]*colB[last]
	comp.InsertLocal(0,
		ifaces.QueryIDf("PIP_Local_%v", ipTracker.GetColID()),
		sym.Sub(column.Shift(ipTracker, -1),
			sym.Mul(column.Shift(colA, -1),
				column.Shift(colB, -1)),
		),
	)

	comp.SubProvers.AppendToInner(0,
		func(run *wizard.ProverRuntime) {
			assignPIP(run, colA, colB, partition, ipTracker)
		},
	)
	return ipTracker
}

// It assigns IPTracker for PIP.
func assignPIP(run *wizard.ProverRuntime, colA, colB, partition, ipTracker ifaces.Column) {
	var (
		cola         = colA.GetColAssignment(run).IntoRegVecSaveAlloc()
		colb         = colB.GetColAssignment(run).IntoRegVecSaveAlloc()
		partitionWit = partition.GetColAssignment(run).IntoRegVecSaveAlloc()
		one          = field.One()
		size         = colA.Size()
	)
	var notPartition field.Element

	var u, v field.Element
	ipTrackerWit := make([]field.Element, size)

	ipTrackerWit[size-1].Mul(&cola[size-1], &colb[size-1])

	for i := size - 2; i >= 0; i-- {
		u.Mul(&cola[i], &colb[i])
		notPartition.Sub(&one, &partitionWit[i+1])
		v.Mul(&ipTrackerWit[i+1], &notPartition)
		ipTrackerWit[i].Add(&u, &v)
	}

	run.AssignColumn(ipTracker.GetColID(), smartvectors.RightZeroPadded(ipTrackerWit, colA.Size()))

}
