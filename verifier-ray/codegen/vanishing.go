package codegen

import (
	"errors"
	"fmt"
	"io"
	"strings"

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
	zw := &zigWriter{w: w}
	zw.line("// Code generated by verifier-ray/testdata/generate; DO NOT EDIT.")
	zw.line("")
	zw.line("const field = @import(\"../field/koalabear.zig\");")
	zw.line("const vanishing = @import(\"../query/vanishing.zig\");")
	zw.line("")

	for i, tc := range cases {
		writeSystemZig(zw, i, tc.Name, tc.System)
	}

	zw.line("pub const systems = [_]vanishing.System{")
	zw.indent++
	for i := range cases {
		zw.linef("system_%d,", i)
	}
	zw.indent--
	zw.line("};")
	return zw.err
}

func writeSystemZig(w *zigWriter, idx int, scenarioName string, sys VanishingSystem) {
	w.linef("// scenario: \"%s\"", zigString(scenarioName))
	w.line("")
	for moduleIdx, module := range sys.Modules {
		if len(module.Buckets) == 1 && len(module.Buckets[0].Vanishings) == 1 {
			w.linef("// expression: \"%s\"", zigString(module.Buckets[0].Vanishings[0].SourceName))
		}
		w.linef("const system_%d_module_%d_expressions = [_]vanishing.ExprNode{", idx, moduleIdx)
		w.indent++
		for _, expr := range module.Expressions {
			writeExprNodeZig(w, expr)
		}
		w.indent--
		w.line("};")
		w.line("")

		for bucketIdx, bucket := range module.Buckets {
			w.linef("const system_%d_module_%d_bucket_%d_vanishings = [_]vanishing.Vanishing{", idx, moduleIdx, bucketIdx)
			w.indent++
			for _, v := range bucket.Vanishings {
				w.linef("// expression: \"%s\"", zigString(v.SourceName))
				w.linef(".{ .expression = %d, .cancelled_positions = &%s },", v.Expression, intSlice(v.CancelledPositions))
			}
			w.indent--
			w.line("};")
			w.line("")
		}

		w.linef("const system_%d_module_%d_buckets = [_]vanishing.Bucket{", idx, moduleIdx)
		w.indent++
		for bucketIdx, bucket := range module.Buckets {
			w.linef(".{ .ratio = %d, .vanishings = &system_%d_module_%d_bucket_%d_vanishings, .quotient_claim_offset = %d },", bucket.Ratio, idx, moduleIdx, bucketIdx, bucket.QuotientClaimOffset)
		}
		w.indent--
		w.line("};")
		w.line("")
	}

	w.linef("const system_%d_modules = [_]vanishing.Module{", idx)
	w.indent++
	for moduleIdx, module := range sys.Modules {
		w.linef("// module: \"%s\"", zigString(module.SourceName))
		w.linef(".{ .size = %s, .expressions = &system_%d_module_%d_expressions, .buckets = &system_%d_module_%d_buckets, .witness_claim_offset = %d },", moduleSizeLiteral(module.Size), idx, moduleIdx, idx, moduleIdx, module.WitnessClaimOffset)
	}
	w.indent--
	w.line("};")
	w.line("")

	w.linef("// system: \"%s\"", zigString(sys.SourceName))
	w.linef("const system_%d = vanishing.System{", idx)
	w.indent++
	w.linef(".modules = &system_%d_modules,", idx)
	w.linef(".dynamic_module_count = %d,", sys.DynamicModuleCount)
	w.linef(".total_witness_claims = %d,", sys.TotalWitnessClaims)
	w.linef(".total_quotient_claims = %d,", sys.TotalQuotientClaims)
	w.indent--
	w.line("};")
	w.line("")
}

func writeExprNodeZig(w *zigWriter, expr ExprNode) {
	switch expr.Kind {
	case ExprColumnClaim:
		w.linef(".{ .column_claim = %d }, // col: \"%s\"", expr.ColumnClaim, zigString(expr.ColumnSourceName))
	case ExprConstant:
		w.linef(".{ .constant = field.Element.init(%d) },", uint64(expr.Constant.Bits()[0]))
	case ExprOp:
		w.linef(".{ .op = .{ .operator = .%s, .operands = &%s } },", expr.Operator, intSlice(expr.Operands))
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

type zigWriter struct {
	w      io.Writer
	indent int
	err    error
}

func (w *zigWriter) line(text string) {
	if w.err != nil {
		return
	}
	_, w.err = fmt.Fprintf(w.w, "%s%s\n", strings.Repeat("    ", w.indent), text)
}

func (w *zigWriter) linef(format string, args ...any) {
	w.line(fmt.Sprintf(format, args...))
}
