package plonk2

import "fmt"

// Curve identifies a scalar field supported by the plonk2 CUDA kernels.
type Curve uint32

const (
	CurveBN254    Curve = 1
	CurveBLS12377 Curve = 2
	CurveBW6761   Curve = 3
)

// CurveInfo records the scalar and base-field widths needed by FFT and MSM
// planning. The modulus arithmetic kernels currently operate on scalar fields.
type CurveInfo struct {
	Curve          Curve
	Name           string
	ScalarLimbs    int
	ScalarBits     int
	BaseFieldLimbs int
}

// Info returns the static parameters associated with c.
func (c Curve) Info() (CurveInfo, bool) {
	switch c {
	case CurveBN254:
		return CurveInfo{
			Curve:          c,
			Name:           "bn254",
			ScalarLimbs:    4,
			ScalarBits:     254,
			BaseFieldLimbs: 4,
		}, true
	case CurveBLS12377:
		return CurveInfo{
			Curve:          c,
			Name:           "bls12-377",
			ScalarLimbs:    4,
			ScalarBits:     253,
			BaseFieldLimbs: 6,
		}, true
	case CurveBW6761:
		return CurveInfo{
			Curve:          c,
			Name:           "bw6-761",
			ScalarLimbs:    6,
			ScalarBits:     377,
			BaseFieldLimbs: 12,
		}, true
	default:
		return CurveInfo{}, false
	}
}

func (c Curve) String() string {
	if info, ok := c.Info(); ok {
		return info.Name
	}
	return fmt.Sprintf("unknown curve %d", c)
}

func (c Curve) validate() (CurveInfo, error) {
	info, ok := c.Info()
	if !ok {
		return CurveInfo{}, fmt.Errorf("plonk2: unsupported curve %d", c)
	}
	return info, nil
}

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}
