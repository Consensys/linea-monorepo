package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	globalcompiler "github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/global"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/wioptest"
)

type globalCompilerCase struct {
	name    string
	modules []globalModuleCase
	honest  globalProofView
	invalid globalProofView
}

type globalModuleCase struct {
	size       int
	exprs      []string
	vanishings []globalVanishingCase
}

type globalVanishingCase struct {
	expr               int
	cancelledPositions []int
}

type globalProofView struct {
	initialRound   runtimeTraceRound
	quotientRound  runtimeTraceRound
	witnessClaims  []field.Ext
	quotientClaims []field.Ext
}

// writeGlobalCompilerCases emits small end-to-end fixtures for the verifier-ray
// global compiler. The proof views are produced by prover-ray's compiler and
// prover actions; verifier-ray replays only the verifier-visible transcript and
// quotient identity.
func writeGlobalCompilerCases(out *bytes.Buffer) {
	builders := []func() *wioptest.VanishingScenario{
		wioptest.NewBooleanColumnVanishingScenario,
		wioptest.NewFibonacciVanishingScenario,
		wioptest.NewPythagoreanTripletVanishingScenario,
	}

	fmt.Fprintln(out, "pub const global_compiler_cases = [_]GlobalCompilerCase{")
	for _, build := range builders {
		writeGlobalCompilerCase(out, buildGlobalCompilerCase(build()))
	}
	fmt.Fprintln(out, "};")
	fmt.Fprintln(out)
}

func buildGlobalCompilerCase(sc *wioptest.VanishingScenario) globalCompilerCase {
	globalcompiler.Compile(sc.Sys)
	exports := collectGlobalVerifierExports(sc.Sys)
	columnIndices := initialColumnIndices(sc.Sys)

	return globalCompilerCase{
		name:    sc.Name,
		modules: buildGlobalModules(exports, columnIndices),
		honest:  runGlobalProof(sc.Sys, sc.AssignHonest, exports),
		invalid: runGlobalProof(sc.Sys, sc.AssignInvalid, exports),
	}
}

func collectGlobalVerifierExports(sys *wiop.System) []globalcompiler.VerifierExport {
	evalRound := sys.Rounds[len(sys.Rounds)-1]
	exports := make([]globalcompiler.VerifierExport, 0, len(evalRound.VerifierActions))
	for _, action := range evalRound.VerifierActions {
		verifier, ok := action.(*globalcompiler.Verifier)
		if !ok {
			continue
		}
		exports = append(exports, verifier.Export())
	}
	return exports
}

func initialColumnIndices(sys *wiop.System) map[*wiop.Column]int {
	result := make(map[*wiop.Column]int)
	for _, col := range sys.Rounds[0].Columns {
		if col.Visibility < wiop.VisibilityOracle {
			continue
		}
		result[col] = len(result)
	}
	return result
}

func buildGlobalModules(exports []globalcompiler.VerifierExport, columnIndices map[*wiop.Column]int) []globalModuleCase {
	modules := make([]globalModuleCase, len(exports))
	for i, export := range exports {
		if export.Module.Size() == 0 {
			panic("global compiler vector generator only supports statically sized modules")
		}

		exprs := &exprLiteralBuilder{columnIndices: columnIndices}
		vanishings := make([]globalVanishingCase, 0, len(export.Module.Vanishings))
		for _, vanishing := range export.Module.Vanishings {
			if vanishing.IsReduced() {
				continue
			}
			vanishings = append(vanishings, globalVanishingCase{
				expr:               exprs.add(vanishing.Expression),
				cancelledPositions: append([]int(nil), vanishing.CancelledPositions...),
			})
		}

		modules[i] = globalModuleCase{
			size:       export.Module.Size(),
			exprs:      exprs.nodes,
			vanishings: vanishings,
		}
	}
	return modules
}

type exprLiteralBuilder struct {
	columnIndices map[*wiop.Column]int
	nodes         []string
}

func (b *exprLiteralBuilder) add(expr wiop.Expression) int {
	switch e := expr.(type) {
	case *wiop.ColumnView:
		columnIndex, ok := b.columnIndices[e.Column]
		if !ok {
			panic(fmt.Sprintf("column %q is not part of the initial verifier-visible round", e.Column.Context.Path()))
		}
		return b.appendNode(fmt.Sprintf(
			".{ .column_view = .{ .column = %d, .shift = %d } }",
			columnIndex,
			e.ShiftingOffset,
		))
	case *wiop.Constant:
		return b.appendNode(fmt.Sprintf(".{ .constant = %d }", u(e.Value)))
	case *wiop.ArithmeticOperation:
		operands := make([]int, len(e.Operands))
		for i, operand := range e.Operands {
			operands[i] = b.add(operand)
		}
		return b.appendNode(fmt.Sprintf(
			".{ .op = .{ .operator = .%s, .operands = &%s } }",
			globalOperatorName(e.Operator),
			intSlice(operands),
		))
	default:
		panic(fmt.Sprintf("unsupported global expression leaf %T", expr))
	}
}

func (b *exprLiteralBuilder) appendNode(literal string) int {
	idx := len(b.nodes)
	b.nodes = append(b.nodes, literal)
	return idx
}

func globalOperatorName(op wiop.ArithmeticOperator) string {
	switch op {
	case wiop.ArithmeticOperatorAdd:
		return "add"
	case wiop.ArithmeticOperatorMul:
		return "mul"
	case wiop.ArithmeticOperatorSub:
		return "sub"
	case wiop.ArithmeticOperatorDiv:
		return "div"
	case wiop.ArithmeticOperatorDouble:
		return "double"
	case wiop.ArithmeticOperatorSquare:
		return "square"
	case wiop.ArithmeticOperatorNegate:
		return "negate"
	case wiop.ArithmeticOperatorInverse:
		return "inverse"
	default:
		panic(fmt.Sprintf("unknown arithmetic operator %v", op))
	}
}

func runGlobalProof(
	sys *wiop.System,
	assign func(rt *wiop.Runtime),
	exports []globalcompiler.VerifierExport,
) globalProofView {
	rt := wiop.NewRuntime(sys)
	assign(&rt)

	for rt.CurrentRound().ID < len(sys.Rounds)-1 {
		rt.AdvanceRound()
		for _, action := range rt.CurrentRound().ProverActions {
			action.Run(rt)
		}
	}

	return globalProofView{
		initialRound:   traceRoundFromRuntime(rt, sys.Rounds[0]),
		quotientRound:  traceRoundFromRuntime(rt, sys.Rounds[len(sys.Rounds)-2]),
		witnessClaims:  collectWitnessClaims(rt, exports),
		quotientClaims: collectQuotientClaims(rt, exports),
	}
}

func traceRoundFromRuntime(rt wiop.Runtime, round *wiop.Round) runtimeTraceRound {
	columns := make([]runtimeTraceColumn, 0, len(round.Columns))
	for _, col := range round.Columns {
		if col.Visibility < wiop.VisibilityOracle {
			continue
		}
		assignment := rt.GetColumnAssignment(col)
		traceColumn := runtimeTraceColumn{
			visibility: col.Visibility,
			assigned:   true,
			isExt:      !assignment.Plain.IsBase(),
		}
		if col.Visibility == wiop.VisibilityOracle {
			if assignment.Plain.IsBase() {
				traceColumn.commitments = commitmentBlocksFromElements(assignment.Plain.AsBase())
			} else {
				traceColumn.commitments = commitmentBlocksFromExts(assignment.Plain.AsExt())
			}
		} else if assignment.Plain.IsBase() {
			traceColumn.baseValues = append([]field.Element(nil), assignment.Plain.AsBase()...)
		} else {
			traceColumn.isExt = true
			traceColumn.extValues = append([]field.Ext(nil), assignment.Plain.AsExt()...)
		}
		columns = append(columns, traceColumn)
	}

	cells := make([]runtimeTraceCell, 0, len(round.Cells))
	for _, cell := range round.Cells {
		value := rt.GetCellValue(cell)
		traceCell := runtimeTraceCell{
			assigned: true,
			isExt:    !value.IsBase(),
		}
		if value.IsBase() {
			traceCell.baseValue = value.AsBase()
		} else {
			traceCell.extValue = value.AsExt()
		}
		cells = append(cells, traceCell)
	}

	return runtimeTraceRound{
		columns: columns,
		cells:   cells,
	}
}

func collectWitnessClaims(rt wiop.Runtime, exports []globalcompiler.VerifierExport) []field.Ext {
	var claims []field.Ext
	for _, export := range exports {
		for _, claim := range export.WitnessClaims {
			claims = append(claims, rt.GetCellValue(claim).AsExt())
		}
	}
	return claims
}

func collectQuotientClaims(rt wiop.Runtime, exports []globalcompiler.VerifierExport) []field.Ext {
	var claims []field.Ext
	for _, export := range exports {
		for _, bucket := range export.Buckets {
			for _, claim := range bucket.QuotientClaims {
				claims = append(claims, rt.GetCellValue(claim).AsExt())
			}
		}
	}
	return claims
}

func writeGlobalCompilerCase(out *bytes.Buffer, tc globalCompilerCase) {
	fmt.Fprintf(out, "    .{ .name = %q,\n", tc.name)
	fmt.Fprintln(out, "        .modules = &.{")
	for _, module := range tc.modules {
		writeGlobalModule(out, module)
	}
	fmt.Fprintln(out, "        },")
	fmt.Fprintln(out, "        .honest =")
	writeGlobalProofView(out, tc.honest)
	fmt.Fprintln(out, "        ,")
	fmt.Fprintln(out, "        .invalid =")
	writeGlobalProofView(out, tc.invalid)
	fmt.Fprintln(out, "    },")
}

func writeGlobalModule(out *bytes.Buffer, module globalModuleCase) {
	fmt.Fprintf(out, "            .{ .size = %d,\n", module.size)
	fmt.Fprintln(out, "                .expressions = &.{")
	for _, expr := range module.exprs {
		fmt.Fprintf(out, "                    %s,\n", expr)
	}
	fmt.Fprintln(out, "                },")
	fmt.Fprintln(out, "                .vanishings = &.{")
	for _, vanishing := range module.vanishings {
		fmt.Fprintf(out,
			"                    .{ .expression = %d, .cancelled_positions = &%s },\n",
			vanishing.expr,
			intSlice(vanishing.cancelledPositions),
		)
	}
	fmt.Fprintln(out, "                },")
	fmt.Fprintln(out, "            },")
}

func writeGlobalProofView(out *bytes.Buffer, proof globalProofView) {
	fmt.Fprintln(out, "            .{")
	fmt.Fprintln(out, "                .initial_round =")
	writeGlobalRoundView(out, proof.initialRound)
	fmt.Fprintln(out, "                ,")
	fmt.Fprintln(out, "                .quotient_round =")
	writeGlobalRoundView(out, proof.quotientRound)
	fmt.Fprintln(out, "                ,")
	fmt.Fprintf(out, "                .witness_claims = &%s,\n", extSlice(proof.witnessClaims))
	fmt.Fprintf(out, "                .quotient_claims = &%s,\n", extSlice(proof.quotientClaims))
	fmt.Fprintln(out, "            }")
}

func writeGlobalRoundView(out *bytes.Buffer, round runtimeTraceRound) {
	fmt.Fprintln(out, "            .{")
	fmt.Fprintln(out, "                .columns = &.{")
	for _, column := range round.columns {
		writeRuntimeTraceColumn(out, column)
	}
	fmt.Fprintln(out, "                },")
	fmt.Fprintln(out, "                .cells = &.{")
	for _, cell := range round.cells {
		writeRuntimeTraceCell(out, cell)
	}
	fmt.Fprintln(out, "                },")
	fmt.Fprintln(out, "            }")
}

func intSlice(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = fmt.Sprintf("%d", value)
	}
	return ".{ " + strings.Join(parts, ", ") + " }"
}
