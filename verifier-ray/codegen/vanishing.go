package codegen

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/global"
)

// UnsupportedExpressionError reports an expression leaf or operation that the
// first verifier-ray vanishing checker intentionally does not evaluate yet.
type UnsupportedExpressionError struct {
	Type string
}

func (e *UnsupportedExpressionError) Error() string {
	return fmt.Sprintf("unsupported vanishing expression %s", e.Type)
}

func IsUnsupportedExpression(err error) bool {
	var unsupported *UnsupportedExpressionError
	return errors.As(err, &unsupported)
}

type NamedVanishingSystem struct {
	Name   string
	System VanishingSystem
}

type VanishingSystem struct {
	SourceName          string
	Modules             []VanishingModule
	DynamicModuleCount  int
	TotalWitnessClaims  int
	TotalQuotientClaims int
}

type ModuleSize struct {
	Dynamic      bool
	StaticSize   int
	DynamicIndex int
}

type VanishingModule struct {
	SourceName         string
	Size               ModuleSize
	Expressions        []ExprNode
	Buckets            []VanishingBucket
	WitnessClaimOffset int
}

type VanishingBucket struct {
	Ratio               int
	Vanishings          []Vanishing
	QuotientClaimOffset int
}

type Vanishing struct {
	SourceName         string
	Expression         int
	CancelledPositions []int
}

type ExprNode struct {
	Kind             ExprKind
	ColumnClaim      int
	ColumnSourceName string
	Constant         field.Element
	Operator         Operator
	Operands         []int
}

type ExprKind int

const (
	ExprColumnClaim ExprKind = iota
	ExprConstant
	ExprOp
)

type Operator string

const (
	OperatorAdd     Operator = "add"
	OperatorMul     Operator = "mul"
	OperatorSub     Operator = "sub"
	OperatorDiv     Operator = "div"
	OperatorDouble  Operator = "double"
	OperatorSquare  Operator = "square"
	OperatorNegate  Operator = "negate"
	OperatorInverse Operator = "inverse"
)

type viewKey struct {
	id    wiop.ObjectID
	shift int
}

// BuildVanishingSystem extracts only compiled global.Verifier actions from sys
// and converts them to the compact data representation consumed by Zig.
func BuildVanishingSystem(sys *wiop.System) (VanishingSystem, error) {
	out := VanishingSystem{SourceName: sys.Context.Path()}
	dynamicIndices := map[*wiop.Module]int{}

	for _, round := range sys.Rounds {
		for _, action := range round.VerifierActions {
			verifier, ok := action.(*global.Verifier)
			if !ok {
				continue
			}
			exp := verifier.Export()
			module := VanishingModule{SourceName: exp.Module.Context.Label, WitnessClaimOffset: out.TotalWitnessClaims}
			if exp.ModuleSize.Dynamic {
				idx, ok := dynamicIndices[exp.Module]
				if !ok {
					idx = len(dynamicIndices)
					dynamicIndices[exp.Module] = idx
				}
				module.Size = ModuleSize{Dynamic: true, DynamicIndex: idx}
			} else {
				module.Size = ModuleSize{StaticSize: exp.ModuleSize.StaticSize}
			}

			views := make(map[viewKey]int, len(exp.WitnessViews))
			for i, view := range exp.WitnessViews {
				views[viewKey{id: view.Column.Context.ID, shift: view.ShiftingOffset}] = i
			}

			out.TotalWitnessClaims += len(exp.WitnessClaims)
			for _, bucket := range exp.Buckets {
				b := VanishingBucket{
					Ratio:               bucket.Ratio,
					QuotientClaimOffset: out.TotalQuotientClaims,
				}
				out.TotalQuotientClaims += len(bucket.QuotientClaims)

				for _, v := range bucket.Vanishings {
					exprIdx, err := appendExpr(&module, views, v.Expression)
					if err != nil {
						return VanishingSystem{}, err
					}
					b.Vanishings = append(b.Vanishings, Vanishing{
						SourceName:         v.Context().Label,
						Expression:         exprIdx,
						CancelledPositions: append([]int(nil), v.CancelledPositions...),
					})
				}
				module.Buckets = append(module.Buckets, b)
			}

			out.Modules = append(out.Modules, module)
		}
	}

	out.DynamicModuleCount = len(dynamicIndices)
	return out, nil
}

func appendExpr(module *VanishingModule, views map[viewKey]int, expr wiop.Expression) (int, error) {
	switch e := expr.(type) {
	case *wiop.ColumnView:
		idx, ok := views[viewKey{id: e.Column.Context.ID, shift: e.ShiftingOffset}]
		if !ok {
			return 0, fmt.Errorf("column view %s shift %d was not exported as a witness claim", e.Column.Context.Path(), e.ShiftingOffset)
		}
		module.Expressions = append(module.Expressions, ExprNode{Kind: ExprColumnClaim, ColumnClaim: idx, ColumnSourceName: e.Column.Context.Label})
		return len(module.Expressions) - 1, nil
	case *wiop.Constant:
		module.Expressions = append(module.Expressions, ExprNode{Kind: ExprConstant, Constant: e.Value})
		return len(module.Expressions) - 1, nil
	case *wiop.ArithmeticOperation:
		operands := make([]int, len(e.Operands))
		for i, operand := range e.Operands {
			idx, err := appendExpr(module, views, operand)
			if err != nil {
				return 0, err
			}
			operands[i] = idx
		}
		op, err := mapOperator(e.Operator)
		if err != nil {
			return 0, err
		}
		module.Expressions = append(module.Expressions, ExprNode{Kind: ExprOp, Operator: op, Operands: operands})
		return len(module.Expressions) - 1, nil
	case *wiop.Cell:
		return 0, &UnsupportedExpressionError{Type: "Cell"}
	case *wiop.CoinField:
		return 0, &UnsupportedExpressionError{Type: "CoinField"}
	default:
		return 0, &UnsupportedExpressionError{Type: fmt.Sprintf("%T", expr)}
	}
}

func mapOperator(op wiop.ArithmeticOperator) (Operator, error) {
	switch op {
	case wiop.ArithmeticOperatorAdd:
		return OperatorAdd, nil
	case wiop.ArithmeticOperatorMul:
		return OperatorMul, nil
	case wiop.ArithmeticOperatorSub:
		return OperatorSub, nil
	case wiop.ArithmeticOperatorDiv:
		return OperatorDiv, nil
	case wiop.ArithmeticOperatorDouble:
		return OperatorDouble, nil
	case wiop.ArithmeticOperatorSquare:
		return OperatorSquare, nil
	case wiop.ArithmeticOperatorNegate:
		return OperatorNegate, nil
	case wiop.ArithmeticOperatorInverse:
		return OperatorInverse, nil
	default:
		return "", &UnsupportedExpressionError{Type: fmt.Sprintf("ArithmeticOperator(%d)", int(op))}
	}
}

// WriteVanishingScenariosZig writes generated Zig source that contains the
// vanishing.System values for the supplied cases. It emits data only; Zig owns
// the evaluator and quotient identity implementation.
func WriteVanishingScenariosZig(w io.Writer, cases []NamedVanishingSystem) error {
	data := vanishingTemplateData{Cases: make([]vanishingTemplateCase, len(cases))}
	for i, tc := range cases {
		data.Cases[i] = newVanishingTemplateCase(i, tc)
	}

	tmpl, err := template.New("vanishing_scenarios").Funcs(template.FuncMap{
		"expr":       exprNodeLiteral,
		"ints":       intSlice,
		"moduleSize": moduleSizeLiteral,
		"zig":        zigString,
	}).Parse(vanishingScenariosTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

type vanishingTemplateData struct {
	Cases []vanishingTemplateCase
}

type vanishingTemplateCase struct {
	Index   int
	Name    string
	System  VanishingSystem
	Modules []vanishingTemplateModule
}

type vanishingTemplateModule struct {
	Index   int
	Module  VanishingModule
	Buckets []vanishingTemplateBucket
}

func (m vanishingTemplateModule) SingleExpressionName() string {
	if len(m.Buckets) != 1 || len(m.Buckets[0].Bucket.Vanishings) != 1 {
		return ""
	}
	return m.Buckets[0].Bucket.Vanishings[0].SourceName
}

type vanishingTemplateBucket struct {
	ModuleIndex int
	Index       int
	Bucket      VanishingBucket
}

func newVanishingTemplateCase(index int, tc NamedVanishingSystem) vanishingTemplateCase {
	out := vanishingTemplateCase{
		Index:   index,
		Name:    tc.Name,
		System:  tc.System,
		Modules: make([]vanishingTemplateModule, len(tc.System.Modules)),
	}
	for moduleIndex, module := range tc.System.Modules {
		moduleView := vanishingTemplateModule{
			Index:   moduleIndex,
			Module:  module,
			Buckets: make([]vanishingTemplateBucket, len(module.Buckets)),
		}
		for bucketIndex, bucket := range module.Buckets {
			moduleView.Buckets[bucketIndex] = vanishingTemplateBucket{
				ModuleIndex: moduleIndex,
				Index:       bucketIndex,
				Bucket:      bucket,
			}
		}
		out.Modules[moduleIndex] = moduleView
	}
	return out
}

const vanishingScenariosTemplate = `// Code generated by verifier-ray/testdata/generate; DO NOT EDIT.

const field = @import("../field/koalabear.zig");
const vanishing = @import("../query/vanishing.zig");

{{range .Cases}}{{- $case := . }}
// scenario: "{{zig $case.Name}}"

{{range $case.Modules}}{{- $module := . }}{{with $module.SingleExpressionName}}// expression: "{{zig .}}"
{{end}}const system_{{$case.Index}}_module_{{$module.Index}}_expressions = [_]vanishing.ExprNode{
{{range $module.Module.Expressions}}    {{expr .}}
{{end}}};

{{range $module.Buckets}}{{- $bucket := . }}const system_{{$case.Index}}_module_{{$bucket.ModuleIndex}}_bucket_{{$bucket.Index}}_vanishings = [_]vanishing.Vanishing{
{{range $bucket.Bucket.Vanishings}}    // expression: "{{zig .SourceName}}"
    .{ .expression = {{.Expression}}, .cancelled_positions = &{{ints .CancelledPositions}} },
{{end}}};

{{end}}const system_{{$case.Index}}_module_{{$module.Index}}_buckets = [_]vanishing.Bucket{
{{range $module.Buckets}}    .{ .ratio = {{.Bucket.Ratio}}, .vanishings = &system_{{$case.Index}}_module_{{.ModuleIndex}}_bucket_{{.Index}}_vanishings, .quotient_claim_offset = {{.Bucket.QuotientClaimOffset}} },
{{end}}};

{{end}}const system_{{$case.Index}}_modules = [_]vanishing.Module{
{{range $case.Modules}}    // module: "{{zig .Module.SourceName}}"
    .{ .size = {{moduleSize .Module.Size}}, .expressions = &system_{{$case.Index}}_module_{{.Index}}_expressions, .buckets = &system_{{$case.Index}}_module_{{.Index}}_buckets, .witness_claim_offset = {{.Module.WitnessClaimOffset}} },
{{end}}};

// system: "{{zig $case.System.SourceName}}"
const system_{{$case.Index}} = vanishing.System{
    .modules = &system_{{$case.Index}}_modules,
    .dynamic_module_count = {{$case.System.DynamicModuleCount}},
    .total_witness_claims = {{$case.System.TotalWitnessClaims}},
    .total_quotient_claims = {{$case.System.TotalQuotientClaims}},
};

{{end}}pub const systems = [_]vanishing.System{
{{range .Cases}}    system_{{.Index}},
{{end}}};
`

func exprNodeLiteral(expr ExprNode) string {
	switch expr.Kind {
	case ExprColumnClaim:
		return fmt.Sprintf(".{ .column_claim = %d }, // col: \"%s\"", expr.ColumnClaim, zigString(expr.ColumnSourceName))
	case ExprConstant:
		return fmt.Sprintf(".{ .constant = field.Element.init(%d) },", uint64(expr.Constant.Bits()[0]))
	case ExprOp:
		return fmt.Sprintf(".{ .op = .{ .operator = .%s, .operands = &%s } },", expr.Operator, intSlice(expr.Operands))
	default:
		panic(fmt.Sprintf("unknown ExprKind %d", int(expr.Kind)))
	}
}

func moduleSizeLiteral(size ModuleSize) string {
	if size.Dynamic {
		return fmt.Sprintf(".{ .dynamic = %d }", size.DynamicIndex)
	}
	return fmt.Sprintf(".{ .static = %d }", size.StaticSize)
}

func intSlice(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = fmt.Sprintf("%d", value)
	}
	return ".{ " + strings.Join(parts, ", ") + " }"
}

func zigString(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(value)
}
