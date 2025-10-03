package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// FromIntVecCoinPositionAccessor[T] refers to a position of a random coin of type
// [coin.IntegerVec].
type FromIntVecCoinPositionAccessor[T zk.Element] struct {
	// Info points to the underlying coin on which the accessor points to.
	Info coin.Info[T]
	// Pos indexes the pointed position in the coin.
	Pos int
}

func (c *FromIntVecCoinPositionAccessor[T]) IsBase() bool {
	//TODO implement me
	panic("implement me")
}

func (c *FromIntVecCoinPositionAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (c *FromIntVecCoinPositionAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	//TODO implement me
	panic("implement me")
}

func (c *FromIntVecCoinPositionAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (c *FromIntVecCoinPositionAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], circ ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// NewFromIntegerVecCoinPosition constructs an [ifaces.Accessor] object refering
// to a specific position of a coin of the [coin.IntegerVec]. It is used to build
// generic verifier columns.
func NewFromIntegerVecCoinPosition[T zk.Element](info coin.Info[T], pos int) ifaces.Accessor[T] {
	if info.Type != coin.IntegerVec {
		panic("expected an coin.IntegerVec")
	}
	if info.Size <= pos {
		utils.Panic("the coin has size %v, but requested position %v", info.Size, pos)
	}
	return &FromIntVecCoinPositionAccessor[T]{
		Info: info,
		Pos:  pos,
	}
}

// Name implements [ifaces.Accessor]
func (c *FromIntVecCoinPositionAccessor[T]) Name() string {
	return fmt.Sprintf("INT_VEC_COIN_ACCESSOR_%v_%v", c.Info.Name, c.Pos)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (c *FromIntVecCoinPositionAccessor[T]) String() string {
	return c.Name()
}

// GetVal implements [ifaces.Accessor]
func (c *FromIntVecCoinPositionAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	return field.NewElement(uint64(run.GetRandomCoinIntegerVec(c.Info.Name)[c.Pos]))
}

// GetFrontendVariable implements [ifaces.Accessor]
func (c *FromIntVecCoinPositionAccessor[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	return circ.GetRandomCoinIntegerVec(c.Info.Name)[c.Pos]
}

// AsVariable implements the [ifaces.Accessor] interface
func (c *FromIntVecCoinPositionAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](c)
}

// Round implements the [ifaces.Accessor] interface
func (c *FromIntVecCoinPositionAccessor[T]) Round() int {
	return c.Info.Round
}
