package symbolic

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"

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
type Variable[T zk.Element] struct {
	Metadata Metadata
}

// Degree implements the [Operator] interface. Yet, this panics if this is called.
func (Variable[T]) Degree([]int) int {
	panic("we never call it for a variable")
}

// Evaluate implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable[T]) Evaluate([]sv.SmartVector) sv.SmartVector {
	panic("we never call it for variables")
}

// EvaluateExt implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable[T]) EvaluateExt([]sv.SmartVector) sv.SmartVector {
	panic("we never call it for variables")
}

func (v Variable[T]) EvaluateMixed([]sv.SmartVector) sv.SmartVector {
	panic("we never call it for variables")
}

// GnarkEval implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable[T]) GnarkEval(api frontend.API, inputs []T) T {
	panic("we never call it for variables")
}

// GnarkEval implements the [Operator] interface. Yet, this panics if this is called.
func (v Variable[T]) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen[T]) gnarkfext.E4Gen[T] {
	panic("we never call it for variables")
}

// NewVariable constructs a new variable object from a parameter implementing
// the [Metadata] interface.
func NewVariable[T zk.Element](metadata Metadata) *Expression[T] {
	return &Expression[T]{
		Children: []*Expression[T]{},
		ESHash:   metadataToESH(metadata),
		Operator: Variable[T]{Metadata: metadata},
		IsBase:   metadata.IsBase(),
	}
}

// metadataToESH gets the ESH from a metadata. It is obtained by hashing
// the string representation of the metadata.
func metadataToESH(m Metadata) fext.GenericFieldElem {
	// since we use bloack2b to hash the string, it is enough to do that on a base field element esh
	var esh field.Element
	sigSeed := []byte(m.String())
	hasher, _ := blake2b.New256(nil)
	hasher.Write(sigSeed)
	sigBytes := hasher.Sum(nil)
	esh.SetBytes(sigBytes)
	// the base field element is then wrapped into an extension element
	return fext.NewESHashFromBase(esh)
}

/*
StringVar is an implementation of [Metadata] aimed toward testing.
*/
type StringVar string

func NewDummyVar[T zk.Element](s string) *Expression[T] {
	return NewVariable[T](StringVar(s))
}

func (s StringVar) String() string { return string(s) }

/*
Validate implements the [Operator] interface.
*/
func (v Variable[T]) Validate(expr *Expression[T]) error {
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

/*
StringVarExt is an implementation of [Metadata] aimed toward testing, but
StringVarExt will be over field extensions.
*/
type StringVarExt string

func NewDummyVarExt[T zk.Element](s string) *Expression[T] {
	return NewVariable[T](StringVarExt(s))
}

func (s StringVarExt) String() string { return string(s) }

/*
IsBase is always false for the StringVarExt testing struct,
as it is considered over field extensions
*/
func (s StringVarExt) IsBase() bool { return false }
