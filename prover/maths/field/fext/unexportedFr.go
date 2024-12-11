package fext

import "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

var rSquare = fr.Element{
	2726216793283724667,
	14712177743343147295,
	12091039717619697043,
	81024008013859129,
}

const (
	q0 uint64 = 725501752471715841
	q1 uint64 = 6461107452199829505
	q2 uint64 = 6968279316240510977
	q3 uint64 = 1345280370688173398
)

// toMont converts z to Montgomery form
// sets and returns z = z * r²
func toMont(z *fr.Element) *fr.Element {
	return z.Mul(z, &rSquare)
}

func smallerThanModulus(z *fr.Element) bool {
	return (z[3] < q3 || (z[3] == q3 && (z[2] < q2 || (z[2] == q2 && (z[1] < q1 || (z[1] == q1 && (z[0] < q0)))))))
}
