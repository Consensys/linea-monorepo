package plonk2

func defaultMSMWindowBits(info CurveInfo, count int) int {
	if info.Curve != CurveBW6761 {
		return 16
	}
	switch {
	case count >= 1<<22:
		return 18
	case count >= 1<<18:
		return 16
	default:
		return 13
	}
}
