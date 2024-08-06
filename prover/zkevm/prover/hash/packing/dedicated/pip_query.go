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

partition := (0,..,1,0,...,1,...,0,..1),

where it is 1 on positions k+1,l+1,n.
Note that a partition ends with 1.


colA := (a0,..,ak,b(k+1),...,bl,c(l+1),..cn)

colB := (a'0,..,a'k,b'(k+1),...,b'l,c'(l+1),..c'n)


Then ipTracker is computed as follows;
 -  ipTracker[last] = colA[last]*colB[last]
 -  ipTracker[i]= colA[i]*colB[i] + ipTracker[i+1]*(1-partition[i])

 Where i stands for the row-index.


 The result of PIP is stored in i-th row of ipTracker where partition[i-1] == 1.

 Thus at position (k,l,n-1) we respectively have (\sum_i a_i*a'_i,\sum_i b_i*b'_i,\sum_i c_i*c'_i)
*/
func InsertPartitionedIP(
	comp *wizard.CompiledIOP,
	name string,
	colA, colB ifaces.Column,
	partition ifaces.Column,
) ifaces.Column {

	ipTracker := comp.InsertCommit(0, ifaces.ColIDf("%v_%v", name, "IPTracker"), colA.Size())
	one := sym.NewConstant(1)

	// Compute the partitioned inner-product
	cola := ifaces.ColumnAsVariable(colA)
	colb := ifaces.ColumnAsVariable(colB)
	shiftTracker := ifaces.ColumnAsVariable(column.Shift(ipTracker, 1))

	// iptaker[i] = (colA[i] * colB[i]) + ipTracker[i+1]* (1-partition[i]).
	expr1 := ifaces.ColumnAsVariable(ipTracker).
		Sub((cola.Mul(colb)).
			Add(shiftTracker.Mul(one.Sub(ifaces.ColumnAsVariable(partition)))),
		)
	comp.InsertGlobal(0, ifaces.QueryIDf("PIP_%v", ipTracker.GetColID()), expr1)

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
		notPartition.Sub(&one, &partitionWit[i])
		v.Mul(&ipTrackerWit[i+1], &notPartition)
		ipTrackerWit[i].Add(&u, &v)
	}

	run.AssignColumn(ipTracker.GetColID(), smartvectors.RightZeroPadded(ipTrackerWit, colA.Size()))

}
