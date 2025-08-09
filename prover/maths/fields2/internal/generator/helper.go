package main

import (
	"fmt"
	"go/types"
	"path"
	"reflect"
	"strings"
)

const (
	InvalidKind = iota
	BasicKind
	SliceKind
	ArrayKind
	NamedKind
	PointerKind
	MapKind
)

// FuncParams holds information about a method parameter to instantiate our copy
// type.
type FuncParams struct {
	Name    string
	SrcType *Type
	NewType *Type
}

func (p *FuncParams) IsMapped() bool {
	return !reflect.DeepEqual(p.SrcType, p.NewType)
}

// Type is a local renderable version of a type implemented as union
type Type struct {
	Kind  int
	Name  string
	Pkg   string
	Elem1 *Type
	Elem2 *Type
	Len   int
}

// NewLocalNamed returns a new local type with the provided name
func NewLocalNamed(name string) *Type {
	return &Type{Kind: NamedKind, Name: name}
}

// NewFromTypesPkg returns a new type from a [types.Type] and a [packages.Package]
func NewFromTypesPkg(typ types.Type) *Type {
	switch typ := typ.(type) {
	case *types.Named:

		if typ.Obj().Pkg() == nil {
			return &Type{
				Kind: NamedKind,
				Name: typ.Obj().Name(),
			}
		}

		return &Type{
			Kind: NamedKind,
			Name: typ.Obj().Name(),
			Pkg:  typ.Obj().Pkg().Path(),
		}
	case *types.Pointer:
		return &Type{
			Kind:  PointerKind,
			Elem1: NewFromTypesPkg(typ.Elem()),
		}
	case *types.Slice:
		return &Type{
			Kind:  SliceKind,
			Elem1: NewFromTypesPkg(typ.Elem()),
		}
	case *types.Array:
		return &Type{
			Kind:  ArrayKind,
			Elem1: NewFromTypesPkg(typ.Elem()),
			Len:   int(typ.Len()),
		}
	case *types.Basic:
		return &Type{
			Kind: BasicKind,
			Name: types.TypeString(typ, nil),
		}
	case *types.Map:
		return &Type{
			Kind:  MapKind,
			Elem1: NewFromTypesPkg(typ.Key()),
			Elem2: NewFromTypesPkg(typ.Elem()),
		}
	case *types.Interface:
		if typ.Empty() {
			return &Type{
				Kind: BasicKind,
				Name: "interface{}",
			}
		}
		panic(fmt.Sprintf("cannot manage referencing interfaces that are not empty: %v", typ.String()))
	case *types.Chan:
		return &Type{
			Kind: BasicKind,
			Name: types.TypeString(typ, nil),
		}
	default:
		panic(fmt.Sprintf("unexpected type: %v", typ.String()))
	}
}

func NewMappedType(typ types.Type, mapping []TypeMapping) *Type {
	switch typ := typ.(type) {
	case *types.Named:

		if typ.Obj().Pkg() == nil {
			return NewLocalNamed(typ.Obj().Name())
		}

		var (
			tName    = typ.Obj().Name()
			tPkgPath = typ.Obj().Pkg().Path()
		)

		for _, m := range mapping {
			if m.SrcType == tName && m.SrcPkg == tPkgPath {
				return NewLocalNamed(m.NewType)
			}
		}

		return NewFromTypesPkg(typ)
	case *types.Pointer:
		return &Type{
			Kind:  PointerKind,
			Elem1: NewMappedType(typ.Elem(), mapping),
		}
	case *types.Slice:
		return &Type{
			Kind:  SliceKind,
			Elem1: NewMappedType(typ.Elem(), mapping),
		}
	case *types.Array:
		return &Type{
			Kind:  ArrayKind,
			Elem1: NewMappedType(typ.Elem(), mapping),
			Len:   int(typ.Len()),
		}
	case *types.Basic:
		return &Type{
			Kind: BasicKind,
			Name: types.TypeString(typ, nil),
		}
	case *types.Map:
		return &Type{
			Kind:  MapKind,
			Elem1: NewMappedType(typ.Key(), mapping),
			Elem2: NewMappedType(typ.Elem(), mapping),
		}
	case *types.Interface:
		if typ.Empty() {
			return &Type{
				Kind: BasicKind,
				Name: "interface{}",
			}
		}
		panic(fmt.Sprintf("cannot manage referencing interfaces that are not empty: %v", typ.String()))
	case *types.Chan:
		return &Type{
			Kind: BasicKind,
			Name: types.TypeString(typ, nil),
		}
	default:
		panic(fmt.Sprintf("unexpected type: %v", typ.String()))
	}
}

// Render returns a string representation of the type used in the template
func (t *Type) Render() string {

	switch t.Kind {
	case BasicKind:
		return t.Name
	case SliceKind:
		return fmt.Sprintf("[]%s", t.Elem1.Render())
	case ArrayKind:
		return fmt.Sprintf("[%d]%s", t.Len, t.Elem1.Render())
	case NamedKind:
		if len(t.Pkg) == 0 {
			return t.Name
		}
		return fmt.Sprintf("%s.%s", path.Base(t.Pkg), t.Name)
	case PointerKind:
		return fmt.Sprintf("*%s", t.Elem1.Render())
	case MapKind:
		return fmt.Sprintf("map[%s]%s", t.Elem1.Render(), t.Elem2.Render())
	default:
		panic("unexpected type")
	}
}

// RenderMethodDefinition renders a method definition
func RenderMethodDefinition(methodInfo MethodInfo) string {

	builder := &strings.Builder{}

	// This renders a comment for the method
	fmt.Fprintf(builder, "// %s is an automatically generated wrapper arround [%v.%v]\n", methodInfo.MethodName, methodInfo.Receiver.SrcType.Render(), methodInfo.MethodName)

	// This renders the signature of the function
	builder.WriteString("func ")
	RenderParamList(builder, []FuncParams{methodInfo.Receiver})
	builder.WriteString(" ") // the space before the method name
	builder.WriteString(methodInfo.MethodName)
	RenderParamList(builder, methodInfo.Params)
	RenderResult(builder, methodInfo.Results)

	// The renders the body of the function
	builder.WriteString(" {\n")

	// This renders the unsafe casts of the parameters
	nbCasted := 0

	if methodInfo.Receiver.IsMapped() {
		builder.WriteString("\t// Casting the parameters using unsafe\n")
		nbCasted++
		RenderUnsafeCast(builder, methodInfo.Receiver)
	}

	for _, p := range methodInfo.Params {
		if !p.IsMapped() {
			continue
		}
		if nbCasted == 0 {
			builder.WriteString("\t// Casting the parameters using unsafe\n")
		}
		nbCasted++
		RenderUnsafeCast(builder, p)
	}

	builder.WriteString("\n")

	// This renders the call to the wrapped function
	builder.WriteString("\t")
	RenderCall(builder, methodInfo)

	builder.WriteString("\n")

	// This renders the unsafe backward cast
	nbCastedCastedBackward := 0
	for i, r := range methodInfo.Results {
		if !r.IsMapped() {
			continue
		}
		if nbCastedCastedBackward == 0 {
			builder.WriteString("\t// Casting the results using unsafe\n")
		}
		nbCastedCastedBackward++
		RenderUnsafeCastBackward(builder, r, i)
	}

	RenderReturnStatement(builder, methodInfo.Results)

	builder.WriteString("}\n")
	return builder.String()
}

// RenderParamList renders a list of parameters enclosed between parenthesis
// without any prefix of suffix space. The parameters are separated by a comma
// followed by a space. This function is used both to render the params of a
// function and the receivers of a method.
func RenderParamList(builder *strings.Builder, params []FuncParams) {
	builder.WriteString("(")
	for i := range params {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(params[i].Name)
		builder.WriteString(" ")
		builder.WriteString(params[i].NewType.Render())
	}
	builder.WriteString(")")
}

// RenderResult renders the result of a function. It writes everything from the
// closing of the parameters declaration parenthesis to the space before the
// opening of the body curly braces.
func RenderResult(builder *strings.Builder, results []FuncParams) {
	switch len(results) {
	case 0:
		// nothing as the result is void and the function declaration would
		// be this way.
		// `func F(...)|_{...}`
		return
	case 1:
		// The structure of the return declaration is as follows:
		// `func F(...)_T|_{ ... }`
		builder.WriteString(" ")
		builder.WriteString(results[0].NewType.Render())
	default:
		// The structure of the return declaration is as follows:
		// `func F(...)_(T1,_T2)|_{ ... }`
		builder.WriteString(" (")
		for i := range results {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(results[i].NewType.Render())
		}
		builder.WriteString(")")
	}
}

// RenderUnsafeCast renders a line of the form
//
// case (ptrType):
// _name := (*SrcType).(unsafe.Pointer(name))
//
// case (else):
// _name := *(*[]SrcType).(unsafe.Pointer(&name))
func RenderUnsafeCast(builder *strings.Builder, params FuncParams) {
	if !params.IsMapped() {
		panic(fmt.Sprintf("calling the unsafe cast on an unmapped type: %+v", params.SrcType))
	}
	if params.NewType.Kind == PointerKind {
		fmt.Fprintf(builder, "\t_%s := (%s)(unsafe.Pointer(%s))\n", params.Name, params.SrcType.Render(), params.Name)
		return
	}
	fmt.Fprintf(builder, "\t_%s := *(*%s)(unsafe.Pointer(&%s))\n", params.Name, params.SrcType.Render(), params.Name)
}

// RenderUnsafeCastBackward is as [RenderUnsafeCast] but converts back to the new type.
func RenderUnsafeCastBackward(builder *strings.Builder, params FuncParams, index int) {
	if !params.IsMapped() {
		panic(fmt.Sprintf("calling the unsafe cast on an unmapped type: %+v", params.SrcType))
	}
	if params.NewType.Kind == PointerKind {
		fmt.Fprintf(builder, "\tr%v := (%s)(unsafe.Pointer(_r%v))\n", index, params.NewType.Render(), index)
		return
	}
	fmt.Fprintf(builder, "\tr%v := *(*%s)(unsafe.Pointer(&_r%v))\n", index, params.NewType.Render(), index)
}

// RenderCall renders the call to the wrapped function
func RenderCall(builder *strings.Builder, methodInfo MethodInfo) {

	// This renders the assignment of the result if this needs being done. The
	// results of the function are unnamed so we assign new variables named
	// _r0 if this is a mapped result that we will need to unsafe-cast backward
	// or r0 if we won't need to unsafe-cast backward.
	if len(methodInfo.Results) > 0 {
		for i := range methodInfo.Results {
			if i > 0 {
				builder.WriteString(", ")
			}
			if methodInfo.Results[i].IsMapped() {
				builder.WriteString("_")
			}
			fmt.Fprintf(builder, "r%v", i)
		}
		builder.WriteString(" := ")
	}

	receiverName := methodInfo.Receiver.Name
	if methodInfo.Receiver.IsMapped() {
		receiverName = "_" + receiverName
	}

	fmt.Fprintf(builder, "%s.%s(", receiverName, methodInfo.MethodName)

	for i := range methodInfo.Params {
		if i > 0 {
			builder.WriteString(", ")
		}

		if methodInfo.Params[i].IsMapped() {
			builder.WriteString("_")
		}
		builder.WriteString(methodInfo.Params[i].Name)
	}

	builder.WriteString(")\n")
}

func RenderReturnStatement(builder *strings.Builder, params []FuncParams) {
	builder.WriteString("\treturn ")
	for i := range params {
		if i > 0 {
			builder.WriteString(", ")
		}
		fmt.Fprintf(builder, "r%v", i)
	}
	builder.WriteString("\n")
}
