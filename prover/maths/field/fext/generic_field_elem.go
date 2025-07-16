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
	Base   field.Element
	Ext    Element
	IsBase bool
}

func NewESHashFromBase(base *field.Element) *GenericFieldElem {
	return &GenericFieldElem{
		Base:   *new(field.Element).Set(base),
		Ext:    NewFromBase(*base),
		IsBase: true,
	}
}

func (e *GenericFieldElem) GetBase() (field.Element, error) {
	if e.IsBase {
		return e.Base, nil
	} else {
		return field.Zero(), fmt.Errorf("cannot get base element from an extension element")
	}
}

func (e *GenericFieldElem) GetExt() Element {
	if e.IsBase {
		e.Ext.SetFromBase(&e.Base)
		return e.Ext
	} else {
		return e.Ext
	}
}

func NewESHashFromExt(ext *Element) *GenericFieldElem {
	return &GenericFieldElem{
		Ext:    *new(Element).Set(ext),
		IsBase: false,
	}
}

func NewMinimalESHashFromExt(ext *Element) *GenericFieldElem {
	if ext.IsBase() {
		baseElem, _ := ext.GetBase()
		return &GenericFieldElem{
			Base:   *new(field.Element).Set(&baseElem),
			Ext:    *new(Element).Set(ext),
			IsBase: true,
		}
	} else {
		return &GenericFieldElem{
			Base:   field.Zero(),
			Ext:    *new(Element).Set(ext),
			IsBase: false,
		}
	}

}

func (e *GenericFieldElem) GetIsBase() bool {
	return e.IsBase
}

func (e *GenericFieldElem) IsEqual(inp *GenericFieldElem) bool {
	// first check if both are actual base elements
	if e.Ext.IsBase() && inp.Ext.IsBase() {
		return e.Ext.A0.Equal(&inp.Ext.A0)
	} else {
		return e.Ext.Equal(&inp.Ext)
	}
}

func (e *GenericFieldElem) IsEqualBase(inp *field.Element) bool {
	if e.IsBase {
		return e.Base.Equal(inp)
	}
	if e.Ext.IsBase() {
		actualBase, _ := e.Ext.GetBase()
		return actualBase.Equal(inp)
	}
	return false
}

func (e *GenericFieldElem) IsEqualExt(inp *Element) bool {
	return e.Ext.Equal(inp)
}

func (e *GenericFieldElem) Set(inp *GenericFieldElem) *GenericFieldElem {
	e.IsBase = inp.IsBase
	e.Ext.Set(&inp.Ext)
	e.Base.Set(&inp.Base)
	return e
}

func (e *GenericFieldElem) Mul(inp *GenericFieldElem) *GenericFieldElem {
	if e.IsBase && inp.IsBase {
		e.Base.Mul(&e.Base, &inp.Base)
		e.IsBase = true
		e.Ext.SetFromBase(&e.Base)
		return e
	} else {
		// not both are base elements
		if e.IsBase {
			e.Ext.MulByBase(&inp.Ext, &e.Base)
		}
		if inp.IsBase {
			e.Ext.MulByBase(&e.Ext, &inp.Base)
		}
		if !e.IsBase && !inp.IsBase {
			// both are extensions
			e.Ext.Mul(&e.Ext, &inp.Ext)
		}

		// Check if the final result is a base element
		if e.Ext.IsBase() {
			actualBase, _ := e.Ext.GetBase()
			e.Base.Set(&actualBase)
			e.IsBase = true
		} else {
			e.Base.SetZero()
			e.IsBase = false
		}
		return e
	}
}

func (e *GenericFieldElem) MulByBase(inp *field.Element) *GenericFieldElem {
	if e.IsBase {
		e.Base.Mul(&e.Base, inp)
		e.Ext.SetFromBase(&e.Base)
	} else {
		// e is an extension
		e.Ext.MulByBase(&e.Ext, inp)
	}
	return e
}

func (e *GenericFieldElem) MulByExt(inp *Element) *GenericFieldElem {
	if e.IsBase {
		e.Ext.MulByBase(inp, &e.Base)
		// now we can set the base to be 0, as it was already used
		e.Base.SetZero()
		e.IsBase = false
	} else {
		// e is an extension
		e.Ext.Mul(&e.Ext, inp)
	}
	return e
}

func (e *GenericFieldElem) Add(inp *GenericFieldElem) *GenericFieldElem {
	if e.IsBase && inp.IsBase {
		e.Base.Add(&e.Base, &inp.Base)
		e.Ext.SetFromBase(&e.Base)
		e.IsBase = true
	} else {
		// not both are base elements
		if e.IsBase {
			e.Ext.AddByBase(&inp.Ext, &e.Base)
		}
		if inp.IsBase {
			e.Ext.AddByBase(&e.Ext, &inp.Base)
		}
		if !e.IsBase && !inp.IsBase {
			// both are extensions
			e.Ext.Add(&e.Ext, &inp.Ext)
		}

		// Check if the final result is a base element
		if e.Ext.IsBase() {
			actualBase, _ := e.Ext.GetBase()
			e.Base.Set(&actualBase)
			e.IsBase = true
		} else {
			e.Base.SetZero()
			e.IsBase = false
		}
	}
	return e
}

func (e *GenericFieldElem) Div(inp *GenericFieldElem) *GenericFieldElem {
	if e.IsBase && inp.IsBase {
		e.Base.Div(&e.Base, &inp.Base)
		e.Ext.SetFromBase(&e.Base)
		e.IsBase = true
	} else {
		// not both are base elements
		if e.IsBase {
			e.Ext.DivByBase(&inp.Ext, &e.Base)
		}
		if inp.IsBase {
			e.Ext.DivByBase(&e.Ext, &inp.Base)
		}
		if !e.IsBase && !inp.IsBase {
			// both are extensions
			e.Ext.Div(&e.Ext, &inp.Ext)
		}

		// Check if the final result is a base element
		if e.Ext.IsBase() {
			actualBase, _ := e.Ext.GetBase()
			e.Base.Set(&actualBase)
			e.IsBase = true
		} else {
			e.Base.SetZero()
			e.IsBase = false
		}
	}
	return e
}

func (e *GenericFieldElem) String() string {
	if e.IsBase {
		return e.Base.String()
	}
	return e.Ext.String()
}

func (e *GenericFieldElem) IsZero() bool {
	if e.IsBase {
		return e.Base.IsZero()
	} else {
		return e.Ext.IsZero()
	}
}

func (e *GenericFieldElem) IsOne() bool {
	if e.IsBase {
		return e.Base.IsOne()
	} else {
		return e.Ext.IsOne()
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
	if inp.IsBase {
		// inp is a base element
		e.Base.Square(&inp.Base)
		e.Ext.SetFromBase(&e.Base)
		e.IsBase = true
		return e
	} else {
		// inp is an extension
		e.Ext.Square(&inp.Ext)
		if e.Ext.IsBase() {
			actualBase, _ := e.Ext.GetBase()
			e.Base.Set(&actualBase)
			e.IsBase = true
		} else {
			e.Base.SetZero()
			e.IsBase = false
		}
		return e
	}
}

func (e *GenericFieldElem) Exp(inp *GenericFieldElem, exponent *big.Int) *GenericFieldElem {
	if inp.IsBase {
		e.Base.Exp(inp.Base, exponent)
		e.Ext.SetFromBase(&e.Base)
		return e
	} else {
		e.Ext.Exp(inp.Ext, exponent)
		if e.Ext.IsBase() {
			actualBase, _ := e.Ext.GetBase()
			e.Base.Set(&actualBase)
			e.IsBase = true
		} else {
			e.Base.SetZero()
			e.IsBase = false
		}
		return e
	}
}

func (e *GenericFieldElem) SetInt64(v int64) *GenericFieldElem {
	e.Base.SetInt64(v)
	e.Ext.SetFromBase(&e.Base)
	e.IsBase = true
	return e
}

func (e *GenericFieldElem) Text(base int) string {
	if e.IsBase {
		return e.Base.Text(base)
	} else {
		return e.Ext.Text(base)
	}
}

func (z *GenericFieldElem) Bytes() []byte {
	if z.GetIsBase() {
		res := z.Base.Bytes()
		return res[:]
	} else {
		res := z.Ext.Bytes()
		return res[:]
	}
}

func (z *GenericFieldElem) Inverse(x *GenericFieldElem) *GenericFieldElem {
	if x.GetIsBase() {
		var resBase field.Element
		resBase.Inverse(&x.Base)
		return NewESHashFromBase(&resBase)
	} else {
		var resExt Element
		resExt.Inverse(&x.Ext)
		return NewESHashFromExt(&resExt)
	}
}
