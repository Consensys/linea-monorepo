package plonk2

import "fmt"

type nttOrder uint8

const (
	nttOrderNatural nttOrder = iota + 1
	nttOrderBitReversed
)

type nttResidency uint8

const (
	nttResidencyHost nttResidency = iota + 1
	nttResidencyDevice
)

type nttDirection uint8

const (
	nttDirectionForward nttDirection = iota + 1
	nttDirectionInverse
	nttDirectionCosetForward
	nttDirectionCosetInverse
)

// NTTPlan records the current transform order and residency contract.
type NTTPlan struct {
	Curve           Curve
	Size            int
	Direction       nttDirection
	InputOrder      nttOrder
	OutputOrder     nttOrder
	InputResidency  nttResidency
	OutputResidency nttResidency
	BatchCount      int
}

func defaultNTTPlan(curve Curve, size int, direction nttDirection, batchCount int) (NTTPlan, error) {
	if _, err := curve.validate(); err != nil {
		return NTTPlan{}, err
	}
	if !isPowerOfTwo(size) {
		return NTTPlan{}, fmt.Errorf("plonk2: NTT size must be a positive power of two")
	}
	if batchCount <= 0 {
		return NTTPlan{}, fmt.Errorf("plonk2: NTT batch count must be positive")
	}
	plan := NTTPlan{
		Curve:           curve,
		Size:            size,
		Direction:       direction,
		InputResidency:  nttResidencyDevice,
		OutputResidency: nttResidencyDevice,
		BatchCount:      batchCount,
	}
	switch direction {
	case nttDirectionForward:
		plan.InputOrder = nttOrderNatural
		plan.OutputOrder = nttOrderBitReversed
	case nttDirectionInverse:
		plan.InputOrder = nttOrderBitReversed
		plan.OutputOrder = nttOrderNatural
	case nttDirectionCosetForward, nttDirectionCosetInverse:
		plan.InputOrder = nttOrderNatural
		plan.OutputOrder = nttOrderNatural
	default:
		return NTTPlan{}, fmt.Errorf("plonk2: unsupported NTT direction %d", direction)
	}
	return plan, nil
}
