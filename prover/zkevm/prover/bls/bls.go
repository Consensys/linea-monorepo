package bls

type group int

const (
	G1 group = iota
	G2
)

const ROUND_NR = 0

func (g group) String() string {
	switch g {
	case G1:
		return "G1"
	case G2:
		return "G2"
	default:
		panic("unknown group")
	}
}
