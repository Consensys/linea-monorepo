package codegen_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/global"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZigVerifier_FibonacciColumnAssignment(t *testing.T) {
	sys, _ := compiledFibSystem(t)
	actions := collectVerifierActions(sys)
	require.Len(t, actions, 1, "compiledFibSystem should produce one verifier action")
	require.IsType(t, &global.Verifier{}, actions[0], "the verifier action should be the global quotient verifier")

	generated := generateFibonacciZigVerifier(t, sys, actions)

	want, err := os.ReadFile("fibonacci_verifier_test.zig")
	require.NoError(t, err)
	assert.Equal(t, string(want), generated,
		"checked-in Zig verifier should match the program extracted from compiledFibSystem")

	zigPath, err := exec.LookPath("zig")
	if err != nil {
		t.Skip("zig binary is not installed")
	}

	tmp := t.TempDir()
	zigFile := filepath.Join(tmp, "fibonacci_verifier_test.zig")
	require.NoError(t, os.WriteFile(zigFile, []byte(generated), 0o600))
	copyZigSupportFile(t, tmp, "koalabear_field.zig")

	//nolint:gosec // Test executes the local Zig compiler on generated test source.
	cmd := exec.Command(zigPath, "test", zigFile)
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"ZIG_GLOBAL_CACHE_DIR="+filepath.Join(tmp, "global-cache"),
		"ZIG_LOCAL_CACHE_DIR="+filepath.Join(tmp, "local-cache"),
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zig test failed:\n%s\nsource:\n%s", out, generated)
}

func collectVerifierActions(sys *wiop.System) []wiop.VerifierAction {
	var actions []wiop.VerifierAction
	for _, r := range sys.Rounds {
		actions = append(actions, r.VerifierActions...)
	}
	return actions
}

func copyZigSupportFile(t *testing.T, dstDir, name string) {
	t.Helper()
	src, err := os.ReadFile(name)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dstDir, name), src, 0o600))
}

func generateFibonacciZigVerifier(t *testing.T, sys *wiop.System, actions []wiop.VerifierAction) string {
	t.Helper()
	require.Len(t, sys.Modules, 1, "Fibonacci fixture should have one module")
	require.Len(t, actions, 1, "Fibonacci fixture should have one verifier action")

	mod := sys.Modules[0]
	require.Equal(t, 8, mod.Size(), "Fibonacci module size should stay stable")
	require.Len(t, mod.Vanishings, 1, "Fibonacci fixture should have one vanishing constraint")

	columns := referencedColumns(t, mod)
	require.Len(t, columns, 1, "Fibonacci verifier should expose one witness column")

	columnNames := make(map[*wiop.Column]string, len(columns))
	for i, col := range columns {
		columnNames[col] = fmt.Sprintf("col%d", i)
	}

	params := zigParams(columns)
	args := zigArgs(columns)
	expr := zigExpr(t, mod.Vanishings[0].Expression, columnNames)
	cancelled := normalizedCancelledRows(mod.Vanishings[0], mod.Size())

	var b strings.Builder
	fmt.Fprintf(&b, "const std = @import(\"std\");\n")
	fmt.Fprintf(&b, "const koalabear = @import(\"koalabear_field.zig\");\n\n")
	fmt.Fprintf(&b, "const ModuleSize: usize = %d;\n", mod.Size())
	b.WriteString(`const Field = koalabear.Field;
const f = koalabear.f;

fn shifted(row: usize, offset: i64) usize {
    const n: i64 = @intCast(ModuleSize);
    const row_i: i64 = @intCast(row);
    return @intCast(@mod(row_i + offset, n));
}

`)
	fmt.Fprintf(&b, "fn isVanishing0Cancelled(row: usize) bool {\n")
	if len(cancelled) == 0 {
		b.WriteString("    return false;\n")
	} else {
		parts := make([]string, len(cancelled))
		for i, row := range cancelled {
			parts[i] = fmt.Sprintf("row == %d", row)
		}
		fmt.Fprintf(&b, "    return %s;\n", strings.Join(parts, " or "))
	}
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "fn evalVanishing0(%s, row: usize) Field {\n", params)
	fmt.Fprintf(&b, "    return %s;\n", expr)
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "fn checkVerifierAction0(%s) !void {\n", params)
	b.WriteString(`    var row: usize = 0;
    while (row < ModuleSize) : (row += 1) {
        if (isVanishing0Cancelled(row)) continue;
        if (!evalVanishing0(`)
	b.WriteString(args)
	b.WriteString(`, row).isZero()) {
            return error.VanishingConstraintFailed;
        }
    }
}

`)
	fmt.Fprintf(&b, "pub fn verifyColumnAssignment(%s) bool {\n", params)
	fmt.Fprintf(&b, "    checkVerifierAction0(%s) catch return false;\n", args)
	b.WriteString(`    return true;
}

const honest = [_]Field{ f(1), f(1), f(2), f(3), f(5), f(8), f(13), f(21) };
const invalid = [_]Field{ f(1), f(1), f(2), f(3), f(5), f(8), f(13), f(22) };

test "generated verifier accepts the Fibonacci column assignment" {
    try checkVerifierAction0(&honest);
    try std.testing.expect(verifyColumnAssignment(&honest));
}

test "generated verifier rejects a broken Fibonacci column assignment" {
    try std.testing.expectError(error.VanishingConstraintFailed, checkVerifierAction0(&invalid));
    try std.testing.expect(!verifyColumnAssignment(&invalid));
}
`)

	return b.String()
}

func referencedColumns(t *testing.T, mod *wiop.Module) []*wiop.Column {
	t.Helper()
	seen := make(map[*wiop.Column]struct{})
	for _, v := range mod.Vanishings {
		collectColumnsFromExpr(t, v.Expression, seen)
	}

	columns := make([]*wiop.Column, 0, len(seen))
	for col := range seen {
		columns = append(columns, col)
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Context.ID < columns[j].Context.ID
	})
	return columns
}

func collectColumnsFromExpr(t *testing.T, expr wiop.Expression, seen map[*wiop.Column]struct{}) {
	t.Helper()
	switch e := expr.(type) {
	case *wiop.ColumnView:
		seen[e.Column] = struct{}{}
	case *wiop.ArithmeticOperation:
		for _, child := range e.Operands {
			collectColumnsFromExpr(t, child, seen)
		}
	case *wiop.Constant:
	default:
		require.FailNowf(t, "unsupported expression", "unsupported expression node %T", expr)
	}
}

func normalizedCancelledRows(v *wiop.Vanishing, n int) []int {
	rows := make([]int, 0, len(v.CancelledPositions))
	for _, p := range v.CancelledPositions {
		if p < 0 {
			rows = append(rows, n+p)
		} else {
			rows = append(rows, p)
		}
	}
	sort.Ints(rows)
	return rows
}

func zigParams(columns []*wiop.Column) string {
	params := make([]string, len(columns))
	for i := range columns {
		params[i] = fmt.Sprintf("col%d: *const [ModuleSize]Field", i)
	}
	return strings.Join(params, ", ")
}

func zigArgs(columns []*wiop.Column) string {
	args := make([]string, len(columns))
	for i := range columns {
		args[i] = fmt.Sprintf("col%d", i)
	}
	return strings.Join(args, ", ")
}

func zigExpr(t *testing.T, expr wiop.Expression, columns map[*wiop.Column]string) string {
	t.Helper()
	switch e := expr.(type) {
	case *wiop.ColumnView:
		name, ok := columns[e.Column]
		require.Truef(t, ok, "missing column name for %s", e.Column.Context.Path())
		return fmt.Sprintf("%s[shifted(row, %d)]", name, e.ShiftingOffset)
	case *wiop.Constant:
		return fmt.Sprintf("f(%d)", e.Value.Uint64())
	case *wiop.ArithmeticOperation:
		operands := make([]string, len(e.Operands))
		for i, child := range e.Operands {
			operands[i] = zigExpr(t, child, columns)
		}
		return zigArithmeticExpr(t, e.Operator, operands)
	default:
		require.FailNowf(t, "unsupported expression", "unsupported expression node %T", expr)
		return ""
	}
}

func zigArithmeticExpr(t *testing.T, op wiop.ArithmeticOperator, operands []string) string {
	t.Helper()
	switch op {
	case wiop.ArithmeticOperatorAdd:
		return fmt.Sprintf("%s.add(%s)", operands[0], operands[1])
	case wiop.ArithmeticOperatorMul:
		return fmt.Sprintf("%s.mul(%s)", operands[0], operands[1])
	case wiop.ArithmeticOperatorSub:
		return fmt.Sprintf("%s.sub(%s)", operands[0], operands[1])
	case wiop.ArithmeticOperatorDiv:
		return fmt.Sprintf("%s.div(%s)", operands[0], operands[1])
	case wiop.ArithmeticOperatorDouble:
		return fmt.Sprintf("%s.add(%s)", operands[0], operands[0])
	case wiop.ArithmeticOperatorSquare:
		return fmt.Sprintf("%s.mul(%s)", operands[0], operands[0])
	case wiop.ArithmeticOperatorNegate:
		return fmt.Sprintf("%s.neg()", operands[0])
	case wiop.ArithmeticOperatorInverse:
		return fmt.Sprintf("%s.inv()", operands[0])
	default:
		require.FailNowf(t, "unsupported operator", "unsupported arithmetic operator %s", op)
		return ""
	}
}
