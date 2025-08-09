package main

import (
	"os"
	"testing"
)

func TestGenerator(t *testing.T) {
	err := Generate(os.Stdout, &GeneratorConfig{
		NewPackageName: "field",
		Mapping: []TypeMapping{
			{
				NewType: "Fr",
				SrcType: "Element",
				SrcPkg:  "github.com/consensys/gnark-crypto/field/koalabear",
			},
			{
				NewType: "Fext",
				SrcType: "E4",
				SrcPkg:  "github.com/consensys/gnark-crypto/field/koalabear/extensions",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Fail()
}

func TestCollect(t *testing.T) {

	cfg := GeneratorConfig{
		Mapping: []TypeMapping{
			{
				NewType: "Fr",
				SrcType: "Element",
				SrcPkg:  "github.com/consensys/gnark-crypto/field/koalabear",
			},
			{
				NewType: "Fext",
				SrcType: "E4",
				SrcPkg:  "github.com/consensys/gnark-crypto/field/koalabear/extensions",
			},
		},
	}

	_, err := collectTypesAndMethods(cfg.Mapping)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTemplate(t *testing.T) {

	koalabearType := &Type{
		Name: "Element",
		Pkg:  "koalabear",
		Kind: NamedKind,
	}

	koalabearPtr := &Type{
		Kind:  PointerKind,
		Elem1: koalabearType,
	}

	newFrType := &Type{
		Name: "Fr",
		Pkg:  "",
		Kind: NamedKind,
	}

	newFrPtr := &Type{
		Kind:  PointerKind,
		Elem1: newFrType,
	}

	e4Type := &Type{
		Name: "E4",
		Pkg:  "extensions",
		Kind: NamedKind,
	}

	e4Ptr := &Type{
		Kind:  PointerKind,
		Elem1: e4Type,
	}

	newFextType := &Type{
		Name: "Fext",
		Pkg:  "",
		Kind: NamedKind,
	}

	newFextPtr := &Type{
		Kind:  PointerKind,
		Elem1: newFextType,
	}

	intType := &Type{
		Name: "int",
		Pkg:  "",
		Kind: BasicKind,
	}

	values := &TemplateValues{
		NewPackageName: "field",
		Packages: []PackageInfo{
			{
				Name: "koalabear",
				Path: "github.com/consensys/gnark-crypto/field/koalabear",
			},
			{
				Name: "extensions",
				Path: "github.com/consensys/gnark-crypto/field/koalabear/extensions",
			},
		},
		Types: []TypeMapping{
			{
				SrcType: "Element",
				SrcPkg:  "koalabear",
				NewType: "Fr",
			},
			{
				SrcType: "Element",
				SrcPkg:  "extensions",
				NewType: "Fext",
			},
		},
		Methods: []MethodInfo{
			// func (z *Fr) Add(a, b *Fr) *Fr
			{
				MethodName: "Add",
				Receiver: FuncParams{
					Name:    "z",
					SrcType: koalabearPtr,
					NewType: newFrPtr,
				},
				Params: []FuncParams{
					{
						Name:    "a",
						SrcType: koalabearPtr,
						NewType: newFrPtr,
					},
					{
						Name:    "b",
						SrcType: koalabearPtr,
						NewType: newFrPtr,
					},
				},
				Results: []FuncParams{
					{
						SrcType: koalabearPtr,
						NewType: newFrPtr,
					},
				},
			},
			// func (z *Fext) Add(a, b *Fext) *Fext
			{
				MethodName: "Add",
				Receiver: FuncParams{
					Name:    "z",
					SrcType: e4Ptr,
					NewType: newFextPtr,
				},
				Params: []FuncParams{
					{
						Name:    "a",
						SrcType: e4Ptr,
						NewType: newFextPtr,
					},
					{
						Name:    "b",
						SrcType: e4Ptr,
						NewType: newFextPtr,
					},
				},
				Results: []FuncParams{
					{
						SrcType: e4Ptr,
						NewType: newFextPtr,
					},
				},
			},
			// func (z *Fr) Exp(x Fr, n int) *Fr
			{
				MethodName: "Exp",
				Receiver: FuncParams{
					Name:    "z",
					SrcType: koalabearPtr,
					NewType: newFrPtr,
				},
				Params: []FuncParams{
					{
						Name:    "x",
						SrcType: koalabearType,
						NewType: newFrType,
					},
					{
						Name:    "n",
						SrcType: intType,
						NewType: intType,
					},
				},
				Results: []FuncParams{
					{
						SrcType: koalabearPtr,
						NewType: newFrPtr,
					},
				},
			},
		},
	}

	if err := fillTemplate(os.Stdout, values); err != nil {
		t.Fatal(err)
	}

	t.Fail()
}
