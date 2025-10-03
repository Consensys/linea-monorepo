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

// var _ ifaces.Accessor = &FromCoinAccessor{}

// FromCoinAccessor implements [ifaces.Accessor] and represents the value of a
// [coin.Info[T]] of type [coin.Field]. It is sometime used to supply a coin to
// a function requiring an accessor explicitly. For [github.com/consensys/linea-monorepo/prover/symbolic.Expression]
// this should not be necessary as [coin.Info[T]] already implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata].
type FromCoinAccessor[T zk.Element] struct {
	// Info represents the underlying [coin.Info[T]] being wrapped by the accessor.
	Info coin.Info[T]
}

// NewFromCoin returns an [ifaces.Accessor] object symbolizing a
// [coin.Info[T]]. The supplied [coin.Info[T]] must be of type [coin.Field] or the
// function panics.
func NewFromCoin[T zk.Element](info coin.Info[T]) ifaces.Accessor[T] {
	if info.Type != coin.Field && info.Type != coin.FieldFromSeed && info.Type != coin.FieldExt {
		utils.Panic("NewFromCoin expects a [coin.Field], a [coin.FieldFromSeed] or a [coin.FieldExt] `info`, got `%v`", info.Type)
	}
	return &FromCoinAccessor[T]{
		Info: info,
	}
}

// Name implements [ifaces.Accessor]
func (c *FromCoinAccessor[T]) Name() string {
	return fmt.Sprintf("COIN_AS_ACCESSOR_%v", c.Info.Name)
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (c *FromCoinAccessor[T]) String() string {
	return c.Name()
}

// GetVal implements [ifaces.Accessor]
func (c *FromCoinAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	return run.GetRandomCoinField(c.Info.Name)
}

func (c *FromCoinAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	return run.GetRandomCoinField(c.Info.Name), nil
}

func (c *FromCoinAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	return run.GetRandomCoinFieldExt(c.Info.Name)
}

// GetFrontendVariable implements [ifaces.Accessor]
func (c *FromCoinAccessor[T]) GetFrontendVariable(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) T {
	return circ.GetRandomCoinField(c.Info.Name)
}

func (c *FromCoinAccessor[T]) GetFrontendVariableBase(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) (T, error) {
	return circ.GetRandomCoinField(c.Info.Name), nil
}

func (c *FromCoinAccessor[T]) GetFrontendVariableExt(_ zk.APIGen[T], circ ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	return circ.GetRandomCoinFieldExt(c.Info.Name)
}

// AsVariable implements the [ifaces.Accessor] interface
func (c *FromCoinAccessor[T]) AsVariable() *symbolic.Expression[T] {
	return symbolic.NewVariable[T](c)
}

// Round implements the [ifaces.Accessor] interface
func (c *FromCoinAccessor[T]) Round() int {
	return c.Info.Round
}

func (c *FromCoinAccessor[T]) IsBase() bool {
	return false
}
