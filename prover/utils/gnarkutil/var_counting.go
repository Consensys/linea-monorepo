package gnarkutil

import (
	"reflect"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend/schema"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// CountVariables count the variables of a circuit without compiling it. It returns
// the number of public, secret and internal variables. The circuit should be
// defined for koalabear.
func CountVariables(circ any) (nbPublic, nbSecret int) {
	// tVar holds a reference to the reflect.Type of [zk.WrappedVariable]
	var (
		tVar = reflect.ValueOf(struct{ A zk.WrappedVariable }{}).FieldByName("A").Type()
	)

	s, err := schema.Walk(ecc.BN254.ScalarField(), circ, tVar, nil) //TODO@yao: check if we should plugin field to replace ecc.BN254.ScalarField()
	if err != nil {
		panic(err)
	}

	return s.Public, s.Secret
}
