package plonk2

import (
	"testing"

	"github.com/consensys/gnark/constraint"
	"github.com/stretchr/testify/require"
)

type scheduleBlueprint struct {
	Schedule constraint.GkrProvingSchedule
}

func (scheduleBlueprint) CalldataSize() int { return 0 }

func (scheduleBlueprint) NbConstraints() int { return 0 }

func (scheduleBlueprint) NbOutputs(constraint.Instruction) int { return 0 }

func (scheduleBlueprint) UpdateInstructionTree(
	constraint.Instruction,
	constraint.InstructionTree,
) constraint.Level {
	return 0
}

func TestNormalizeGkrScheduleLevels_DecodedPointerLevels(t *testing.T) {
	skip := constraint.GkrSkipLevel{
		Wires:        []int{1},
		ClaimSources: []constraint.GkrClaimSource{{Level: 2}},
	}
	single := constraint.GkrSingleSourceZeroCheckLevel{
		Wires:        []int{3},
		ClaimSources: []constraint.GkrClaimSource{{Level: 4}},
	}
	sumcheck := constraint.GkrSumcheckLevel{{
		Wires:        []int{5},
		ClaimSources: []constraint.GkrClaimSource{{Level: 6}},
	}}

	blueprint := &scheduleBlueprint{
		Schedule: constraint.GkrProvingSchedule{&skip, &single, &sumcheck},
	}

	normalizeGkrScheduleLevels([]constraint.Blueprint{blueprint})

	require.IsType(t, constraint.GkrSkipLevel{}, blueprint.Schedule[0])
	require.IsType(t, constraint.GkrSingleSourceZeroCheckLevel{}, blueprint.Schedule[1])
	require.IsType(t, constraint.GkrSumcheckLevel{}, blueprint.Schedule[2])
	require.Equal(t, skip, blueprint.Schedule[0])
	require.Equal(t, single, blueprint.Schedule[1])
	require.Equal(t, sumcheck, blueprint.Schedule[2])
}
