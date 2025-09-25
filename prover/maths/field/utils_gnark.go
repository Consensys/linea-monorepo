package field

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

func Exp[T zk.Element](api zk.APIGen[T], a T, n *big.Int) *T {

	res := api.ValueOf(1)
	nBytes := n.Bytes()

	// TODO handle negative case
	for _, b := range nBytes {
		for j := 0; j < 8; j++ {
			c := (b >> (7 - j)) & 1
			res = api.Mul(res, res)
			if c == 1 {
				res = api.Mul(res, &a)
			}
		}
	}
	return res
}
