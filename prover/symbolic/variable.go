package symbolic

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"golang.org/x/crypto/blake2b"
)

// Metadata is an interface that must be implemented by type in order to be
// used to instantiate a [Variable] with them.
type Metadata interface {
	/*
		Strings allows addressing a map by variable 2 instances for which
		String() returns the same result are treated as equal.
	*/
	String() string
	IsBase() bool
}

// Variable implements the [Operator] interface and implements a variable; i.e.
// an abstract value that can be assigned in order to evaluate the expression.
type Variable struct {
	Metadata Metadata
}

// Degree implements the [Operator] interface. Yet, this panics if this is called.
func (Variable) Degree([]int) int {
	panic("we never call it for a variable")
}

// Evaluate implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable) Evaluate([]sv.SmartVector, ...mempool.MemPool) sv.SmartVector {
	panic("we never call it for variables")
}

// EvaluateExt implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable) EvaluateExt([]sv.SmartVector, ...mempool.MemPool) sv.SmartVector {
	panic("we never call it for variables")
}

// GnarkEval implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {
	panic("we never call it for variables")
}

// NewVariable constructs a new variable object from a parameter implementing
// the [Metadata] interface.
func NewVariable(metadata Metadata) *Expression {
	return &Expression{
		Children: []*Expression{},
		ESHash:   metadataToESH(metadata),
		Operator: Variable{Metadata: metadata},
		isBase:   metadata.IsBase(),
	}
}

// metadataToESH gets the ESH from a metadata. It is obtained by hashing
// the string representation of the metadata.
func metadataToESH(m Metadata) fext.Element {
	// since we use bloack2b to hash the string, it is enough to do that on a base field element esh
	var esh field.Element
	sigSeed := []byte(m.String())
	hasher, _ := blake2b.New256(nil)
	hasher.Write(sigSeed)
	sigBytes := hasher.Sum(nil)
	esh.SetBytes(sigBytes)
	// the base field element is then wrapped into an extension element
	return fext.NewFromBase(esh)
}

/*
StringVar is an implementation of [Metadata] aimed toward testing.
*/
type StringVar string

func NewDummyVar(s string) *Expression {
	return NewVariable(StringVar(s))
}

func (s StringVar) String() string { return string(s) }

/*
Validate implements the [Operator] interface.
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

/*
IsBase is always true for the StringVar testing struct
*/
func (s StringVar) IsBase() bool { return true }
