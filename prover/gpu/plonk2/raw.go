package plonk2

import "fmt"

// RawLayout records gnark-crypto raw word counts at the plonk2 API boundary.
type RawLayout struct {
	Curve                Curve
	ScalarWords          int
	AffinePointWords     int
	ProjectivePointWords int
}

// RawLayoutForCurve returns raw uint64 word counts for gnark-crypto layouts.
func RawLayoutForCurve(curve Curve) (RawLayout, error) {
	info, err := curve.validate()
	if err != nil {
		return RawLayout{}, err
	}
	return RawLayout{
		Curve:                curve,
		ScalarWords:          info.ScalarLimbs,
		AffinePointWords:     2 * info.BaseFieldLimbs,
		ProjectivePointWords: 3 * info.BaseFieldLimbs,
	}, nil
}

func validateRawAffinePoints(curve Curve, points []uint64) (RawLayout, int, error) {
	layout, err := RawLayoutForCurve(curve)
	if err != nil {
		return RawLayout{}, 0, err
	}
	if len(points) == 0 {
		return RawLayout{}, 0, fmt.Errorf("plonk2: point buffer must not be empty")
	}
	if len(points)%layout.AffinePointWords != 0 {
		return RawLayout{}, 0, fmt.Errorf(
			"plonk2: point word count %d is not a multiple of %d",
			len(points),
			layout.AffinePointWords,
		)
	}
	return layout, len(points) / layout.AffinePointWords, nil
}

func validateRawScalarsExact(curve Curve, scalars []uint64, count int) error {
	layout, err := RawLayoutForCurve(curve)
	if err != nil {
		return err
	}
	want := count * layout.ScalarWords
	if len(scalars) != want {
		return fmt.Errorf("plonk2: scalar word count %d, want %d", len(scalars), want)
	}
	return nil
}

func validateRawScalarsAtMost(curve Curve, scalars []uint64, maxCount int) (int, error) {
	layout, err := RawLayoutForCurve(curve)
	if err != nil {
		return 0, err
	}
	if len(scalars) == 0 {
		return 0, fmt.Errorf("plonk2: scalar buffer must not be empty")
	}
	if len(scalars)%layout.ScalarWords != 0 {
		return 0, fmt.Errorf(
			"plonk2: scalar word count %d is not a multiple of %d",
			len(scalars),
			layout.ScalarWords,
		)
	}
	count := len(scalars) / layout.ScalarWords
	if count > maxCount {
		return 0, fmt.Errorf("plonk2: scalar count %d exceeds point count %d", count, maxCount)
	}
	return count, nil
}
