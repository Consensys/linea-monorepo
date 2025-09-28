package accessors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// var _ ifaces.Accessor = &FromConstAccessor{}

// FromConstAccessor implements [ifaces.Accessor]. It symbolizes a constant
// value and be used as input to functions that expects an [ifaces.Accessor] but
// where the caller is only interested in passing a constant value.
type FromConstAccessor[T zk.Element] struct {
	// F is the constant served by the accessor
	Base       field.Element
	Ext        fext.Element
	IsBaseFlag bool
}

// NewConstant returns an [ifaces.Accessor] object representing a constant value.
func NewConstant[T zk.Element](f field.Element) ifaces.Accessor[T] {
	return &FromConstAccessor[T]{
		Base:       f,
		Ext:        fext.Lift(f),
		IsBaseFlag: true,
	}
}

// NewConstant returns an [ifaces.Accessor] object representing a constant value.
func NewConstantExt[T zk.Element](f fext.Element) ifaces.Accessor[T] {
	return &FromConstAccessor[T]{
		Base:       field.Zero(),
		Ext:        f,
		IsBaseFlag: false,
	}
}

// Name implements [ifaces.Accessor]
func (c *FromConstAccessor[T]) Name() string {
	if c.IsBaseFlag {
		return fmt.Sprintf("CONST_ACCESSOR_%v", c.Base.String())
	} else {
		return fmt.Sprintf("CONST_ACCESSOR_%v", c.Ext.String())
	}
}

// String implements [github.com/consensys/linea-monorepo/prover/symbolic.Metadata]
func (c *FromConstAccessor[T]) String() string {
	return c.Name()
}

// GetVal implements [ifaces.Accessor]
func (c *FromConstAccessor[T]) GetVal(run ifaces.Runtime) field.Element {
	return c.Base
}

func (c *FromConstAccessor[T]) GetValBase(run ifaces.Runtime) (field.Element, error) {
	if c.IsBaseFlag {
		return c.Base, nil
	} else {
		panic("Requested a base field element from an accessor defined over field extensions.")
	}
}

func (c *FromConstAccessor[T]) GetValExt(run ifaces.Runtime) fext.Element {
	return c.Ext
}

func (c *FromConstAccessor[T]) GetFrontendVariable(api zk.APIGen[T], _ ifaces.GnarkRuntime[T]) T {
	if c.IsBaseFlag {
		return *api.ValueOf(c.Base)
	} else {
		panic("Requested a base field element from an accessor defined over field extensions.")
	}
}

func (c *FromConstAccessor[T]) GetFrontendVariableBase(api zk.APIGen[T], _ ifaces.GnarkRuntime[T]) (T, error) {
	if c.IsBaseFlag {
		return *api.ValueOf(c.Base), nil
	} else {
		panic("Requested a base field element from an accessor defined over field extensions.")
	}
}

func (c *FromConstAccessor[T]) GetFrontendVariableExt(api zk.APIGen[T], _ ifaces.GnarkRuntime[T]) gnarkfext.E4Gen[T] {
	var e gnarkfext.E4Gen[T]
	e4Api, err := gnarkfext.NewExt4[T](api.GnarkAPI())
	if err != nil {
		panic(err)
	}
	e = *e4Api.NewFromExt(c.Ext)
	return e
}

// AsVariable implements the [ifaces.Accessor] interface
func (c *FromConstAccessor[T]) AsVariable() *symbolic.Expression[T] {
	if c.IsBaseFlag {
		return symbolic.NewConstant[T](c.Base)
	} else {
		return symbolic.NewConstant[T](c.Ext)
	}
}

// Round implements the [ifaces.Accessor] interface
func (c *FromConstAccessor[T]) Round() int {
	return 0
}

func (c *FromConstAccessor[T]) IsBase() bool {
	return c.IsBaseFlag
}
