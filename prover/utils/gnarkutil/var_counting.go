package gnarkutil

import (
	"errors"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/schema"
)

// CountVariables count the variables of a circuit without compiling it. It returns
// the number of public, secret and internal variables.
func CountVariables(circ frontend.Circuit) (nbPublic, nbSecret, nbInternal int) {

	// tVar holds a reference to the reflect.Type of [frontend.Variable]
	var (
		_tVar frontend.Variable = 1
		tVar                    = reflect.TypeOf(_tVar)
	)

	// leaf handlers are called when encountering leafs in the circuit data struct
	// leafs are Constraints that need to be initialized in the context of compiling a circuit
	variableCounter := func(f schema.LeafInfo, tInput reflect.Value) error {
		if tInput.CanSet() {

			switch f.Visibility {
			case schema.Unset:
				return errors.New("can't set val " + f.FullName() + " visibility is unset")
			case schema.Public:
				nbPublic++
				return nil
			case schema.Internal:
				nbInternal++
				return nil
			case schema.Secret:
				nbSecret++
				return nil
			}

			return errors.New("unexpected visibility " + string(f.Visibility))
		}
		return errors.New("can't set val " + f.FullName())
	}

	schema.Walk(circ, tVar, variableCounter)

	return nbPublic, nbSecret, nbInternal
}
