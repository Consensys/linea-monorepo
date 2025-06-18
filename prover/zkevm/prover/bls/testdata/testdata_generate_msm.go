package main

import "math/big"

type msmInputType int

const (
	msmScalarTrivial msmInputType = iota // 0
	msmScalarRange                       // scalar is in range of the scalar field
	msmScalarBig                         // scalar is not in range of the scalar field. But it is still a valid scalar

	msmPointTrivial    // point is 0
	msmPointOnCurve    // point is on curve but not in subgroup
	msmPointInSubgroup // point is in subgroup
	msmPointInvalid    // point is not on curve
)

type msmInput[T affine] struct {
	scalar msmInputType
	point  msmInputType

	n *big.Int
	P T

	ToMSMCircuit        bool
	ToGroupCheckCircuit bool
}

type msmInputCase[T affine] struct {
	inputs []msmInput[T]

	Res T
}
