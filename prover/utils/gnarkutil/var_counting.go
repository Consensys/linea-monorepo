package gnarkutil

// CountVariables count the variables of a circuit without compiling it. It returns
// the number of public, secret and internal variables. The circuit should be
// defined for koalabear.
func CountVariables(circ any) (nbPublic, nbSecret int) {

	panic("not implemented to support multiple fields especially koalabear")

	// // tVar holds a reference to the reflect.Type of [T]
	// var (
	// 	tVar = reflect.ValueOf(struct{ A T }{}).FieldByName("A").Type()
	// )

	// s, err := schema.Walk(ecc.BN254.ScalarField(), circ, tVar, nil) //TODO@yao: check if we should plugin field to replace ecc.BN254.ScalarField()
	// if err != nil {
	// 	panic(err)
	// }

	// return s.Public, s.Secret
}
