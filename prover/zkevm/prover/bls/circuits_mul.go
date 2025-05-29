package bls

const (
	nbRowsPerG1Mul = nbFpLimbs + 3*nbG1Limbs // 1 scalar, 1 point, 1 previous accumulator, 1 next accumulator
	nbRowsPerG2Mul = nbFpLimbs + 3*nbG2Limbs // 1 scalar, 1 point, 1 previous accumulator, 1 next accumulator
)
