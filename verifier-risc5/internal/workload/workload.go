package workload

var DefaultWords = [...]uint64{
	0x0123456789abcdef,
	0xfedcba9876543210,
	0x0f1e2d3c4b5a6978,
	0x8877665544332211,
	0x13579bdf2468ace0,
	0xc001d00dcafef00d,
}

const DefaultExpected uint64 = 0xd89160ddca835f9f

func Mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

func Compute(words []uint64) uint64 {
	acc := uint64(0x9e3779b97f4a7c15)

	for _, word := range words {
		acc = Mix64(acc ^ word)
	}

	return acc
}
