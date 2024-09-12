package ecpair

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/fields_bn254"
	"github.com/consensys/gnark/std/algebra/emulated/sw_bn254"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/bitslice"
	"github.com/consensys/gnark/std/math/emulated"
)

var fpParams sw_bn254.BaseField

// G1ElementWizard represents G1 element as Wizard limbs (2 limbs of 128 bits)
type G1ElementWizard struct {
	P [nbG1Limbs]frontend.Variable
}

// ToG1Element converts G1ElementWizard to G1Affine used in circuit
func (c *G1ElementWizard) ToG1Element(api frontend.API, fp *emulated.Field[sw_bn254.BaseField]) sw_bn254.G1Affine {
	PXlimbs := make([]frontend.Variable, fpParams.NbLimbs())
	PYlimbs := make([]frontend.Variable, fpParams.NbLimbs())
	PXlimbs[2], PXlimbs[3] = bitslice.Partition(api, c.P[0], 64, bitslice.WithNbDigits(128))
	PXlimbs[0], PXlimbs[1] = bitslice.Partition(api, c.P[1], 64, bitslice.WithNbDigits(128))
	PYlimbs[2], PYlimbs[3] = bitslice.Partition(api, c.P[2], 64, bitslice.WithNbDigits(128))
	PYlimbs[0], PYlimbs[1] = bitslice.Partition(api, c.P[3], 64, bitslice.WithNbDigits(128))
	PX := fp.NewElement(PXlimbs)
	PY := fp.NewElement(PYlimbs)
	P := sw_bn254.G1Affine{
		X: *PX,
		Y: *PY,
	}
	return P
}

// G2ElementWizard represents G2 element as Wizard limbs (4 limbs of 128 bits)
type G2ElementWizard struct {
	Q [nbG2Limbs]frontend.Variable
}

// ToG2Element converts G2ElementWizard to G2Affine used in circuit
func (c *G2ElementWizard) ToG2Element(api frontend.API, fp *emulated.Field[sw_bn254.BaseField]) sw_bn254.G2Affine {
	QXAlimbs := make([]frontend.Variable, fpParams.NbLimbs())
	QXBlimbs := make([]frontend.Variable, fpParams.NbLimbs())
	QYAlimbs := make([]frontend.Variable, fpParams.NbLimbs())
	QYBlimbs := make([]frontend.Variable, fpParams.NbLimbs())

	// arithmetization provides G2 coordinates in the following order:
	//   X_Im, X_Re, Y_Im, Y_Re
	// but in gnark we expect
	//   X_Re, X_Im, Y_Re, Y_Im
	// so we need to swap the limbs.
	QXBlimbs[2], QXBlimbs[3] = bitslice.Partition(api, c.Q[0], 64, bitslice.WithNbDigits(128))
	QXBlimbs[0], QXBlimbs[1] = bitslice.Partition(api, c.Q[1], 64, bitslice.WithNbDigits(128))
	QXAlimbs[2], QXAlimbs[3] = bitslice.Partition(api, c.Q[2], 64, bitslice.WithNbDigits(128))
	QXAlimbs[0], QXAlimbs[1] = bitslice.Partition(api, c.Q[3], 64, bitslice.WithNbDigits(128))
	QYBlimbs[2], QYBlimbs[3] = bitslice.Partition(api, c.Q[4], 64, bitslice.WithNbDigits(128))
	QYBlimbs[0], QYBlimbs[1] = bitslice.Partition(api, c.Q[5], 64, bitslice.WithNbDigits(128))
	QYAlimbs[2], QYAlimbs[3] = bitslice.Partition(api, c.Q[6], 64, bitslice.WithNbDigits(128))
	QYAlimbs[0], QYAlimbs[1] = bitslice.Partition(api, c.Q[7], 64, bitslice.WithNbDigits(128))

	QXA := fp.NewElement(QXAlimbs)
	QXB := fp.NewElement(QXBlimbs)
	QX := fields_bn254.E2{
		A0: *QXA,
		A1: *QXB,
	}
	QYA := fp.NewElement(QYAlimbs)
	QYB := fp.NewElement(QYBlimbs)
	QY := fields_bn254.E2{
		A0: *QYA,
		A1: *QYB,
	}

	var Q sw_bn254.G2Affine
	Q.P.X = QX
	Q.P.Y = QY

	return Q
}

// GtElementWizard represents Gt element as Wizard limbs (24 limbs of 128 bits)
type GtElementWizard struct {
	T [nbGtLimbs]frontend.Variable
}

// ToGtElement converts GtElementWizard to target group element used in circuit
func (c *GtElementWizard) ToGtElement(api frontend.API, fp *emulated.Field[sw_bn254.BaseField]) sw_bn254.GTEl {
	C0B0XLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C0B0YLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C0B1XLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C0B1YLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C0B2XLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C0B2YLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C1B0XLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C1B0YLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C1B1XLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C1B1YLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C1B2XLimbs := make([]frontend.Variable, fpParams.NbLimbs())
	C1B2YLimbs := make([]frontend.Variable, fpParams.NbLimbs())

	C0B0XLimbs[2], C0B0XLimbs[3] = bitslice.Partition(api, c.T[0], 64, bitslice.WithNbDigits(128))
	C0B0XLimbs[0], C0B0XLimbs[1] = bitslice.Partition(api, c.T[1], 64, bitslice.WithNbDigits(128))
	C0B0YLimbs[2], C0B0YLimbs[3] = bitslice.Partition(api, c.T[2], 64, bitslice.WithNbDigits(128))
	C0B0YLimbs[0], C0B0YLimbs[1] = bitslice.Partition(api, c.T[3], 64, bitslice.WithNbDigits(128))
	C0B1XLimbs[2], C0B1XLimbs[3] = bitslice.Partition(api, c.T[4], 64, bitslice.WithNbDigits(128))
	C0B1XLimbs[0], C0B1XLimbs[1] = bitslice.Partition(api, c.T[5], 64, bitslice.WithNbDigits(128))
	C0B1YLimbs[2], C0B1YLimbs[3] = bitslice.Partition(api, c.T[6], 64, bitslice.WithNbDigits(128))
	C0B1YLimbs[0], C0B1YLimbs[1] = bitslice.Partition(api, c.T[7], 64, bitslice.WithNbDigits(128))
	C0B2XLimbs[2], C0B2XLimbs[3] = bitslice.Partition(api, c.T[8], 64, bitslice.WithNbDigits(128))
	C0B2XLimbs[0], C0B2XLimbs[1] = bitslice.Partition(api, c.T[9], 64, bitslice.WithNbDigits(128))
	C0B2YLimbs[2], C0B2YLimbs[3] = bitslice.Partition(api, c.T[10], 64, bitslice.WithNbDigits(128))
	C0B2YLimbs[0], C0B2YLimbs[1] = bitslice.Partition(api, c.T[11], 64, bitslice.WithNbDigits(128))
	C1B0XLimbs[2], C1B0XLimbs[3] = bitslice.Partition(api, c.T[12], 64, bitslice.WithNbDigits(128))
	C1B0XLimbs[0], C1B0XLimbs[1] = bitslice.Partition(api, c.T[13], 64, bitslice.WithNbDigits(128))
	C1B0YLimbs[2], C1B0YLimbs[3] = bitslice.Partition(api, c.T[14], 64, bitslice.WithNbDigits(128))
	C1B0YLimbs[0], C1B0YLimbs[1] = bitslice.Partition(api, c.T[15], 64, bitslice.WithNbDigits(128))
	C1B1XLimbs[2], C1B1XLimbs[3] = bitslice.Partition(api, c.T[16], 64, bitslice.WithNbDigits(128))
	C1B1XLimbs[0], C1B1XLimbs[1] = bitslice.Partition(api, c.T[17], 64, bitslice.WithNbDigits(128))
	C1B1YLimbs[2], C1B1YLimbs[3] = bitslice.Partition(api, c.T[18], 64, bitslice.WithNbDigits(128))
	C1B1YLimbs[0], C1B1YLimbs[1] = bitslice.Partition(api, c.T[19], 64, bitslice.WithNbDigits(128))
	C1B2XLimbs[2], C1B2XLimbs[3] = bitslice.Partition(api, c.T[20], 64, bitslice.WithNbDigits(128))
	C1B2XLimbs[0], C1B2XLimbs[1] = bitslice.Partition(api, c.T[21], 64, bitslice.WithNbDigits(128))
	C1B2YLimbs[2], C1B2YLimbs[3] = bitslice.Partition(api, c.T[22], 64, bitslice.WithNbDigits(128))
	C1B2YLimbs[0], C1B2YLimbs[1] = bitslice.Partition(api, c.T[23], 64, bitslice.WithNbDigits(128))

	C0B0X := fp.NewElement(C0B0XLimbs)
	C0B0Y := fp.NewElement(C0B0YLimbs)
	C0B1X := fp.NewElement(C0B1XLimbs)
	C0B1Y := fp.NewElement(C0B1YLimbs)
	C0B2X := fp.NewElement(C0B2XLimbs)
	C0B2Y := fp.NewElement(C0B2YLimbs)
	C1B0X := fp.NewElement(C1B0XLimbs)
	C1B0Y := fp.NewElement(C1B0YLimbs)
	C1B1X := fp.NewElement(C1B1XLimbs)
	C1B1Y := fp.NewElement(C1B1YLimbs)
	C1B2X := fp.NewElement(C1B2XLimbs)
	C1B2Y := fp.NewElement(C1B2YLimbs)

	T := sw_bn254.GTEl{
		C0: fields_bn254.E6{
			B0: fields_bn254.E2{
				A0: *C0B0X,
				A1: *C0B0Y,
			},
			B1: fields_bn254.E2{
				A0: *C0B1X,
				A1: *C0B1Y,
			},
			B2: fields_bn254.E2{
				A0: *C0B2X,
				A1: *C0B2Y,
			},
		},
		C1: fields_bn254.E6{
			B0: fields_bn254.E2{
				A0: *C1B0X,
				A1: *C1B0Y,
			},
			B1: fields_bn254.E2{
				A0: *C1B1X,
				A1: *C1B1Y,
			},
			B2: fields_bn254.E2{
				A0: *C1B2X,
				A1: *C1B2Y,
			},
		},
	}
	return T
}

// MultiG2GroupcheckCircuit is a circuit that checks multiple G2 group
// membership. Use [newMultiG2GroupcheckCircuit] to create a new instance with
// bounded number of allowed checks.
type MultiG2GroupcheckCircuit struct {
	Instances []G2GroupCheckInstance `gnark:",public"`
}

func newMultiG2GroupcheckCircuit(nbInstances int) *MultiG2GroupcheckCircuit {
	return &MultiG2GroupcheckCircuit{
		Instances: make([]G2GroupCheckInstance, nbInstances),
	}
}

func (c *MultiG2GroupcheckCircuit) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bn254.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field emulation: %w", err)
	}
	pairing, err := sw_bn254.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instances {
		if err := c.Instances[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("instance %d check: %w", i, err)
		}
	}
	return nil
}

// G2GroupCheckInstance is a single instance of G2 group check.
type G2GroupCheckInstance struct {
	Q         G2ElementWizard
	IsSuccess frontend.Variable
}

func (c *G2GroupCheckInstance) Check(api frontend.API, fp *emulated.Field[sw_bn254.BaseField], pairing *sw_bn254.Pairing) error {
	Q := c.Q.ToG2Element(api, fp)

	evmprecompiles.ECPairIsOnG2(api, &Q, c.IsSuccess)
	return nil
}

// MultiG1GroupcheckCircuit is a circuit that checks multiple Miller loop
// computation correctness. Use [newMultiMillerLoopMulCircuit] to create a new
// instance with bounded number of allowed checks.
type MultiMillerLoopMulCircuit struct {
	Instances []MillerLoopMulInstance `gnark:",public"`
}

func newMultiMillerLoopMulCircuit(nbInstance int) *MultiMillerLoopMulCircuit {
	return &MultiMillerLoopMulCircuit{
		Instances: make([]MillerLoopMulInstance, nbInstance),
	}
}

func (c *MultiMillerLoopMulCircuit) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bn254.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field emulation: %w", err)
	}
	pairing, err := sw_bn254.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instances {
		if err := c.Instances[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("instance %d check: %w", i, err)
		}
	}
	return nil
}

// MillerLoopMulInstance is a single instance of Miller loop computation check.
type MillerLoopMulInstance struct {
	Prev    GtElementWizard
	P       G1ElementWizard
	Q       G2ElementWizard
	Current GtElementWizard
}

func (c *MillerLoopMulInstance) Check(api frontend.API, fp *emulated.Field[sw_bn254.BaseField], pairing *sw_bn254.Pairing) error {
	P := c.P.ToG1Element(api, fp)
	Q := c.Q.ToG2Element(api, fp)
	prev := c.Prev.ToGtElement(api, fp)
	current := c.Current.ToGtElement(api, fp)

	return evmprecompiles.ECPairMillerLoopAndMul(api, &prev, &P, &Q, &current)
}

// MultiMillerLoopFinalExpCircuit is a circuit that checks multiple Miller loop
// and final exponentiation checks. Use [newMultiMillerLoopFinalExpCircuit] to
// create a new instance with bounded number of allowed checks.
type MultiMillerLoopFinalExpCircuit struct {
	Instances []MillerLoopFinalExpInstance `gnark:",public"`
}

func newMultiMillerLoopFinalExpCircuit(nbInstance int) *MultiMillerLoopFinalExpCircuit {
	return &MultiMillerLoopFinalExpCircuit{
		Instances: make([]MillerLoopFinalExpInstance, nbInstance),
	}
}

func (c *MultiMillerLoopFinalExpCircuit) Define(api frontend.API) error {
	fp, err := emulated.NewField[sw_bn254.BaseField](api)
	if err != nil {
		return fmt.Errorf("new field emulation: %w", err)
	}
	pairing, err := sw_bn254.NewPairing(api)
	if err != nil {
		return fmt.Errorf("new pairing: %w", err)
	}
	for i := range c.Instances {
		if err := c.Instances[i].Check(api, fp, pairing); err != nil {
			return fmt.Errorf("instance %d check: %w", i, err)
		}
	}
	return nil
}

// MillerLoopFinalExpInstance is a single instance of Miller loop and final
// exponentiation check.
type MillerLoopFinalExpInstance struct {
	Prev     GtElementWizard
	P        G1ElementWizard
	Q        G2ElementWizard
	Expected [2]frontend.Variable
}

func (c *MillerLoopFinalExpInstance) Check(api frontend.API, fp *emulated.Field[sw_bn254.BaseField], pairing *sw_bn254.Pairing) error {
	P := c.P.ToG1Element(api, fp)
	Q := c.Q.ToG2Element(api, fp)
	prev := c.Prev.ToGtElement(api, fp)
	api.AssertIsEqual(c.Expected[0], 0)

	return evmprecompiles.ECPairMillerLoopAndFinalExpCheck(api, &prev, &P, &Q, c.Expected[1])
}
