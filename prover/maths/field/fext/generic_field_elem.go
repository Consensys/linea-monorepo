package fext

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"math/big"
)

/*
GenericFieldElem is a generic field element that can either be a base field element or an extension field element.
It should only be used in places where performance is not critical, as it incurs overhead due to the
storing both a base and extension field version, as well as checks and conversions between base and extension elements.
*/
type GenericFieldElem struct {
	base   field.Element
	ext    Element
	isBase bool
}

func NewESHashFromBase(base *field.Element) *GenericFieldElem {
	return &GenericFieldElem{
		base:   *new(field.Element).Set(base),
		ext:    NewFromBase(*base),
		isBase: true,
	}
}

func (e *GenericFieldElem) GetBase() (field.Element, error) {
	if e.isBase {
		return e.base, nil
	} else {
		return field.Zero(), fmt.Errorf("cannot get base element from an extension element")
	}
}

func (e *GenericFieldElem) GetExt() Element {
	if e.isBase {
		e.ext.SetFromBase(&e.base)
		return e.ext
	} else {
		return e.ext
	}
}

func NewESHashFromExt(ext *Element) *GenericFieldElem {
	return &GenericFieldElem{
		ext:    *new(Element).Set(ext),
		isBase: false,
	}
}

func NewMinimalESHashFromExt(ext *Element) *GenericFieldElem {
	if ext.IsBase() {
		baseElem, _ := ext.GetBase()
		return &GenericFieldElem{
			base:   *new(field.Element).Set(&baseElem),
			ext:    *new(Element).Set(ext),
			isBase: true,
		}
	} else {
		return &GenericFieldElem{
			base:   field.Zero(),
			ext:    *new(Element).Set(ext),
			isBase: false,
		}
	}

}

func (e *GenericFieldElem) IsBase() bool {
	return e.isBase
}

func (e *GenericFieldElem) IsEqual(inp *GenericFieldElem) bool {
	// first check if both are actual base elements
	if e.ext.IsBase() && inp.ext.IsBase() {
		return e.ext.A0.Equal(&inp.ext.A0)
	} else {
		return e.ext.Equal(&inp.ext)
	}
}

func (e *GenericFieldElem) IsEqualBase(inp *field.Element) bool {
	if e.isBase {
		return e.base.Equal(inp)
	}
	if e.ext.IsBase() {
		actualBase, _ := e.ext.GetBase()
		return actualBase.Equal(inp)
	}
	return false
}

func (e *GenericFieldElem) IsEqualExt(inp *Element) bool {
	return e.ext.Equal(inp)
}

func (e *GenericFieldElem) Set(inp *GenericFieldElem) *GenericFieldElem {
	e.isBase = inp.isBase
	e.ext.Set(&inp.ext)
	e.base.Set(&inp.base)
	return e
}

func (e *GenericFieldElem) Mul(inp *GenericFieldElem) *GenericFieldElem {
	if e.isBase && inp.isBase {
		e.base.Mul(&e.base, &inp.base)
		e.isBase = true
		e.ext.SetFromBase(&e.base)
		return e
	} else {
		// not both are base elements
		if e.isBase {
			e.ext.MulByBase(&inp.ext, &e.base)
		}
		if inp.isBase {
			e.ext.MulByBase(&e.ext, &inp.base)
		}
		if !e.isBase && !inp.isBase {
			// both are extensions
			e.ext.Mul(&e.ext, &inp.ext)
		}

		// Check if the final result is a base element
		if e.ext.IsBase() {
			actualBase, _ := e.ext.GetBase()
			e.base.Set(&actualBase)
			e.isBase = true
		} else {
			e.base.SetZero()
			e.isBase = false
		}
		return e
	}
}

func (e *GenericFieldElem) Add(inp *GenericFieldElem) *GenericFieldElem {
	if e.isBase && inp.isBase {
		e.base.Add(&e.base, &inp.base)
		e.ext.SetFromBase(&e.base)
		e.isBase = true
	} else {
		// not both are base elements
		if e.isBase {
			e.ext.AddByBase(&inp.ext, &e.base)
		}
		if inp.isBase {
			e.ext.AddByBase(&e.ext, &inp.base)
		}
		if !e.isBase && !inp.isBase {
			// both are extensions
			e.ext.Add(&e.ext, &inp.ext)
		}

		// Check if the final result is a base element
		if e.ext.IsBase() {
			actualBase, _ := e.ext.GetBase()
			e.base.Set(&actualBase)
			e.isBase = true
		} else {
			e.base.SetZero()
			e.isBase = false
		}
	}
	return e
}

func (e *GenericFieldElem) Div(inp *GenericFieldElem) *GenericFieldElem {
	if e.isBase && inp.isBase {
		e.base.Div(&e.base, &inp.base)
		e.ext.SetFromBase(&e.base)
		e.isBase = true
	} else {
		// not both are base elements
		if e.isBase {
			e.ext.DivByBase(&inp.ext, &e.base)
		}
		if inp.isBase {
			e.ext.DivByBase(&e.ext, &inp.base)
		}
		if !e.isBase && !inp.isBase {
			// both are extensions
			e.ext.Div(&e.ext, &inp.ext)
		}

		// Check if the final result is a base element
		if e.ext.IsBase() {
			actualBase, _ := e.ext.GetBase()
			e.base.Set(&actualBase)
			e.isBase = true
		} else {
			e.base.SetZero()
			e.isBase = false
		}
	}
	return e
}

func (e *GenericFieldElem) String() string {
	if e.isBase {
		return e.base.String()
	}
	return e.ext.String()
}

func (e *GenericFieldElem) IsZero() bool {
	if e.isBase {
		return e.base.IsZero()
	} else {
		return e.ext.IsZero()
	}
}

func (e *GenericFieldElem) IsOne() bool {
	if e.isBase {
		return e.base.IsOne()
	} else {
		return e.ext.IsOne()
	}
}

func GenericFieldOne() *GenericFieldElem {
	baseOne := field.One()
	return NewESHashFromBase(&baseOne)
}

func GenericFieldZero() *GenericFieldElem {
	baseZero := field.Zero()
	return NewESHashFromBase(&baseZero)
}

func (e *GenericFieldElem) Square(inp *GenericFieldElem) *GenericFieldElem {
	if inp.isBase {
		// inp is a base element
		e.base.Square(&inp.base)
		e.ext.SetFromBase(&e.base)
		e.isBase = true
		return e
	} else {
		// inp is an extension
		e.ext.Square(&inp.ext)
		if e.ext.IsBase() {
			actualBase, _ := e.ext.GetBase()
			e.base.Set(&actualBase)
			e.isBase = true
		} else {
			e.base.SetZero()
			e.isBase = false
		}
		return e
	}
}

func (e *GenericFieldElem) Exp(inp *GenericFieldElem, exponent *big.Int) *GenericFieldElem {
	if inp.isBase {
		e.base.Exp(inp.base, exponent)
		e.ext.SetFromBase(&e.base)
		return e
	} else {
		e.ext.Exp(inp.ext, exponent)
		if e.ext.IsBase() {
			actualBase, _ := e.ext.GetBase()
			e.base.Set(&actualBase)
			e.isBase = true
		} else {
			e.base.SetZero()
			e.isBase = false
		}
		return e
	}
}

func (e *GenericFieldElem) SetInt64(v int64) *GenericFieldElem {
	e.base.SetInt64(v)
	e.ext.SetFromBase(&e.base)
	e.isBase = true
	return e
}

func (e *GenericFieldElem) Text(base int) string {
	if e.isBase {
		return e.base.Text(base)
	} else {
		return e.ext.Text(base)
	}
}

func (z *GenericFieldElem) Bytes() []byte {
	if z.IsBase() {
		res := z.base.Bytes()
		return res[:]
	} else {
		res := z.ext.Bytes()
		return res[:]
	}
}

func (z *GenericFieldElem) Inverse(x *GenericFieldElem) *GenericFieldElem {
	if x.IsBase() {
		var resBase field.Element
		resBase.Inverse(&x.base)
		return NewESHashFromBase(&resBase)
	} else {
		var resExt Element
		resExt.Inverse(&x.ext)
		return NewESHashFromExt(&resExt)
	}
}
