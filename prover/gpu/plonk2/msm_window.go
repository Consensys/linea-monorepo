package plonk2

func defaultMSMWindowBits(info CurveInfo, count int) int {
	if info.Curve == CurveBN254 {
		if count < 1<<13 {
			return 8
		}
		return 16
	}
	if info.Curve != CurveBW6761 {
		return 16
	}
	if count >= 1<<19 {
		return 18
	}
	return 14
}
