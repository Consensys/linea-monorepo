package symbolic

import (
	"fmt"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
	"golang.org/x/crypto/blake2b"
)

/*
Constraint on the type of a variable
*/
type Metadata interface {
	/*
		Strings allows adressing a map by variable 2 instances for which
		String() returns the same result are treated as equal.
	*/
	String() string
}

/*
Implements `Operator“
*/
type Variable struct {
	Metadata Metadata
}

func (Variable) Degree([]int) int {
	panic("we never call it for a variable")
}

func (v Variable) Evaluate([]sv.SmartVector) sv.SmartVector {
	panic("we never call it for variables")
}

func (v Variable) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {
	panic("we never call it for variables")
}

/*
Returns an expression with a unique node
*/
func NewVariable(metadata Metadata) *Expression {
	return &Expression{
		Children: []*Expression{},
		ESHash:   metadataToESH(metadata),
		Operator: Variable{Metadata: metadata},
	}
}

// Gets the ESH from a metadata. It is obtained by hashing
// the string representation of the metadata.
func metadataToESH(m Metadata) field.Element {
	var esh field.Element
	sigSeed := []byte(m.String())
	hasher, _ := blake2b.New256(nil)
	hasher.Write(sigSeed)
	sigBytes := hasher.Sum(nil)
	esh.SetBytes(sigBytes)
	return esh
}

/*
Dummy implementation of metadata. Used for testing internally.
*/
type StringVar string

func NewDummyVar(s string) *Expression {
	return NewVariable(StringVar(s))
}

func (s StringVar) String() string { return string(s) }

/*
Validates that the Variable is well-formed
*/
func (v Variable) Validate(expr *Expression) error {
	// This test that the variable is indeed the operator of the expression it is
	// tested on.
	if !reflect.DeepEqual(v, expr.Operator) {
		utils.Panic("expr.operator %#v != v %#v", v, expr.Operator)
	}

	if len(expr.Children) != 0 {
		return fmt.Errorf("mismatch in the size of the children and coefficients")
	}

	return nil
}
