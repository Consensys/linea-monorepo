package single_round

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

func Prove(run *wizard.ProverRuntime, smallSize, size int) {

	p1 := make([]field.Element, size)
	p2 := make([]field.Element, size)
	p3 := make([]field.Element, size)
	p4 := make([]field.Element, smallSize)

	for i := 0; i < size; i++ {
		if i%2 == 0 {
			p1[i].SetRandom()
		} else {
			p2[(i+3)%size].SetInterface(i % smallSize)
		}
	}

	copy(p3, p2)
	vector.Reverse(p3)

	for i := 0; i < smallSize; i++ {
		p4[i].SetInterface(i)
	}

	run.AssignColumn(P1, smartvectors.NewRegular(p1))
	run.AssignColumn(P2, smartvectors.NewRegular(p2))
	run.AssignColumn(P3, smartvectors.NewRegular(p3))
	run.AssignColumn(P4, smartvectors.NewRegular(p4))

}
