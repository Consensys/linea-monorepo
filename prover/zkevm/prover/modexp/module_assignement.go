package modexp

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"

	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

// antichamberAssignment is a builder structure used to incrementally compute
// the assignment of the column of the [Module] module.
type antichamberAssignment struct {
	isActive    *common.VectorBuilder
	isSmall     *common.VectorBuilder
	isLarge     *common.VectorBuilder
	limbs       [common.NbLimbU128]*common.VectorBuilder
	toSmallCirc *common.VectorBuilder
}

// Assign assigns the anti-chamber module
func (mod *Module) Assign(run *wizard.ProverRuntime) {

	mod.Input.assignIsModexp(run)

	var (
		modexpCountSmall int = 0
		modexpCountLarge int = 0
		isModexp             = mod.Input.IsModExp.GetColAssignment(run).IntoRegVecSaveAlloc()
		settings             = mod.Input.Settings
		builder              = antichamberAssignment{
			isActive:    common.NewVectorBuilder(mod.IsActive),
			isSmall:     common.NewVectorBuilder(mod.IsSmall),
			isLarge:     common.NewVectorBuilder(mod.IsLarge),
			toSmallCirc: common.NewVectorBuilder(mod.ToSmallCirc),
		}
	)

	// Retrieve the limbs assignment and initialize the limb builders
	var limbs [common.NbLimbU128][]field.Element
	for i := range common.NbLimbU128 {
		limbs[i] = mod.Input.Limbs[i].GetColAssignment(run).IntoRegVecSaveAlloc()
		builder.limbs[i] = common.NewVectorBuilder(mod.Limbs[i])
	}

	limbSize := len(limbs[0])
	for currPosition := 0; currPosition < limbSize; {

		if isModexp[currPosition].IsZero() {
			currPosition++
			continue
		}

		// This sanity-check is purely defensive and will indicate that we
		// missed the start of a Modexp instance
		for i := range common.NbLimbU128 {
			if len(limbs[i])-currPosition < modexpNumRowsPerInstance {
				utils.Panic("A new modexp is starting but there is not enough rows (currPosition=%v len(ecdata.Limb)=%v)", currPosition, len(limbs))
			}
		}

		isLarge := false

		// An instance is considered large if any of the operand has more than
		// 2 16-bytes limbs (or 16 2-bytes limbs).
		for k := 0; k < modexpNumRowsPerInstance; k++ {
			isZeroLimbs := true
			for i := range common.NbLimbU128 {
				isZeroLimbs = isZeroLimbs && limbs[i][currPosition+k].IsZero()
			}

			if k%32 < 30 && !isZeroLimbs {
				isLarge = true
				break
			}
		}

		if isLarge {
			modexpCountLarge++
		} else {
			modexpCountSmall++
		}

		for k := 0; k < modexpNumRowsPerInstance; k++ {

			builder.isActive.PushOne()
			builder.isSmall.PushBoolean(!isLarge)
			builder.isLarge.PushBoolean(isLarge)

			for i := range common.NbLimbU128 {
				builder.limbs[i].PushField(limbs[i][currPosition+k])
			}

			if !isLarge && k%32 >= 30 {
				builder.toSmallCirc.PushOne()
			} else {
				builder.toSmallCirc.PushZero()
			}
		}

		currPosition += modexpNumRowsPerInstance
	}

	if modexpCountSmall > settings.MaxNbInstance256 {
		exit.OnLimitOverflow(
			settings.MaxNbInstance256,
			modexpCountSmall,
			fmt.Errorf("limit overflow: the modexp (256 bits) count is %v and the limit is %v", modexpCountSmall, settings.MaxNbInstance256),
		)
	}

	if modexpCountLarge > settings.MaxNbInstance4096 {
		exit.OnLimitOverflow(
			settings.MaxNbInstance4096,
			modexpCountLarge,
			fmt.Errorf("limit overflow: the modexp (4096 bits) count is %v and the limit is %v", modexpCountLarge, settings.MaxNbInstance4096),
		)
	}

	builder.isActive.PadAndAssign(run, field.Zero())
	builder.isSmall.PadAndAssign(run, field.Zero())
	builder.isLarge.PadAndAssign(run, field.Zero())
	builder.toSmallCirc.PadAndAssign(run, field.Zero())
	for i := range common.NbLimbU128 {
		builder.limbs[i].PadAndAssign(run, field.Zero())
	}

	// It is possible to not declare the circuit (for testing purpose) in that
	// case we skip the corresponding assignment part.
	if mod.HasCircuit {
		mod.FlattenLimbsSmall.Run(run)
		mod.FlattenLimbsLarge.Run(run)

		mod.GnarkCircuitConnector256Bits.Assign(run)
		mod.GnarkCircuitConnector4096Bits.Assign(run)
	}
}
