package selector

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/sirupsen/logrus"
)

func SelectIntoMessage(comp *wizard.CompiledIOP, handles []ifaces.Column, selector coin.Info) (res []ifaces.Column) {

	round := selector.Round // the selector must be declared last

	logrus.Warnf("the selector is not constrained %v", selector.Name)

	for _, handle := range handles {
		selectedName := ifaces.ColIDf("SELECTED_%v_%v", handle.GetColID(), selector.Name)
		selected := comp.InsertPublicInput(round, ifaces.ColID(selectedName), selector.Size)
		res = append(res, selected)
	}

	comp.SubProvers.AppendToInner(round, func(assi *wizard.ProverRuntime) {
		// assign the returned messages
		selector := assi.GetRandomCoinIntegerVec(selector.Name)
		for i, handle := range handles {
			witness := handle.GetColAssignment(assi)
			selected := []field.Element{}
			for _, pos := range selector {
				selected = append(selected, witness.Get(pos))
			}
			assi.AssignColumn(res[i].GetColID(), smartvectors.NewRegular(selected))
		}
	})

	return res

}
