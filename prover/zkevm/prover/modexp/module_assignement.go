package modexp

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// antichamberAssignment is a builder structure used to incrementally compute
// the assignment of the column of the [Module] module.
type antichamberAssignment struct {
	isActive    *common.VectorBuilder
	isSmall     *common.VectorBuilder
	isLarge     *common.VectorBuilder
	limbs       *common.VectorBuilder
	toSmallCirc *common.VectorBuilder
}

// Assign assigns the anti-chamber module
func (mod *Module) Assign(run *wizard.ProverRuntime) {

	mod.Input.assignIsModexp(run)

	var (
		isModexp = mod.Input.isModExp.GetColAssignment(run).IntoRegVecSaveAlloc()
		limbs    = mod.Input.Limbs.GetColAssignment(run).IntoRegVecSaveAlloc()
		builder  = antichamberAssignment{
			isActive:    common.NewVectorBuilder(mod.IsActive),
			isSmall:     common.NewVectorBuilder(mod.IsSmall),
			isLarge:     common.NewVectorBuilder(mod.IsLarge),
			limbs:       common.NewVectorBuilder(mod.Limbs),
			toSmallCirc: common.NewVectorBuilder(mod.ToSmallCirc),
		}
	)

	for currPosition := 0; currPosition < len(limbs); {

		if isModexp[currPosition].IsOne() {

			isLarge := false

			for k := 0; k < modexpNumRowsPerInstance; k++ {
				if k%32 < 30 && !limbs[currPosition+k].IsZero() {
					isLarge = true
					break
				}
			}

			for k := 0; k < modexpNumRowsPerInstance; k++ {

				builder.isActive.PushOne()
				builder.isSmall.PushBoolean(!isLarge)
				builder.isLarge.PushBoolean(isLarge)
				builder.limbs.PushField(limbs[currPosition+k])

				if !isLarge && k%32 >= 30 {
					builder.toSmallCirc.PushOne()
				} else {
					builder.toSmallCirc.PushZero()
				}
			}

			currPosition += modexpNumRowsPerInstance
		}

		currPosition++
	}

	builder.isActive.PadAndAssign(run, field.Zero())
	builder.isSmall.PadAndAssign(run, field.Zero())
	builder.isLarge.PadAndAssign(run, field.Zero())
	builder.limbs.PadAndAssign(run, field.Zero())
	builder.toSmallCirc.PadAndAssign(run, field.Zero())

	// It is possible to not declare the circuit (for testing purpose) in that
	// case we skip the corresponding assignment part.
	if mod.hasCircuit {
		mod.GnarkCircuitConnector256Bits.Assign(run)
		mod.GnarkCircuitConnector4096Bits.Assign(run)
	}
}
