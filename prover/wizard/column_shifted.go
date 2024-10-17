package wizard

import (
	"strconv"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

type ColShifted struct {
	parent *ColNatural
	offset int
}

func (shf *ColShifted) String() string {
	return shf.parent.String() + ":<<" + strconv.Itoa(shf.offset)
}

func (shf *ColShifted) GetAssignment(run Runtime) smartvectors.SmartVector {
	return shf.parent.GetAssignment(run).RotateRight(-shf.offset)
}

func (shf *ColShifted) GetAssignmentGnark(api frontend.API, run RuntimeGnark) []frontend.Variable {
	parent := shf.parent.GetAssignmentGnark(api, run) // [a b c d e f g h]
	return shiftFrontendVarSlice(parent, shf.offset)
}

func (shf *ColShifted) Size() int {
	return shf.parent.size
}

func (shf *ColShifted) Round() int {
	return shf.parent.round
}

func (shf *ColShifted) Shift(n int) Column {

	if shf.offset+n == 0 {
		return shf.parent
	}

	return &ColShifted{
		parent: shf.parent,
		offset: n + shf.offset,
	}
}

func shiftFrontendVarSlice(parent []frontend.Variable, n int) []frontend.Variable {
	res := make([]frontend.Variable, len(parent))
	for i := range res {
		posParent := utils.PositiveMod(i-n, len(parent))
		res[i] = parent[posParent]
	}
	return res
}
