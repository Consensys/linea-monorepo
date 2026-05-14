package global_test

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/global"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/wioptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompile_Completeness verifies that for every vanishing scenario, an
// honest witness satisfies all verifier actions that Compile registers.
func TestCompile_Completeness(t *testing.T) {
	for _, build := range wioptest.VanishingScenarios() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			global.Compile(sc.Sys)
			rt := wiop.NewRuntime(sc.Sys)
			sc.AssignHonest(&rt)
			require.NoError(t, wioptest.RunAndVerify(&rt),
				"compiled verifier must accept an honest witness")
		})
	}
}

// TestCompile_Soundness verifies that for every vanishing scenario, an invalid
// witness (one that violates at least one constraint) is rejected by the
// compiled verifier. This is the core soundness property of the compiler: a
// cheating prover cannot produce a quotient that satisfies the identity check.
func TestCompile_Soundness(t *testing.T) {
	for _, build := range wioptest.VanishingScenarios() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			global.Compile(sc.Sys)
			rt := wiop.NewRuntime(sc.Sys)
			sc.AssignInvalid(&rt)
			assert.Error(t, wioptest.RunAndVerify(&rt),
				"compiled verifier must reject an invalid witness")
		})
	}
}

func Test_zigTest(t *testing.T) {
	zigTest(t)
}

func zigTest(t *testing.T) {
	t.Helper()
	zigPath, zigErr := exec.LookPath("zig")

	repoRoot := repoRootFromTest(t)
	outputDir := filepath.Join(repoRoot, "wiop", "zigverifiers")
	zigStatics := filepath.Join(repoRoot, "wiop", "zigstatics", "koalabear_field.zig")
	require.NoError(t, os.MkdirAll(outputDir, 0o755))

	for _, build := range wioptest.VanishingScenarios() {
		sc := build()
		t.Run(sc.Name, func(t *testing.T) {
			global.Compile(sc.Sys)

			rt := wiop.NewRuntime(sc.Sys)
			sc.AssignHonest(&rt)
			require.NoError(t, wioptest.RunAndVerify(&rt),
				"Go verifier must accept before generating the equivalent Zig verifier")

			src, err := findGlobalVerifier(t, sc.Sys).GenerateZig(rt)
			require.NoError(t, err)

			zigFile := filepath.Join(outputDir, zigVerifierFileName(sc.Name))
			require.NoError(t, os.WriteFile(zigFile, src, 0o600))
			if zigErr != nil {
				return
			}

			//nolint:gosec // Test executes the local Zig compiler on generated verifier source.
			cmd := exec.Command(zigPath, "test", "--dep", "koalabear", "-Mroot="+zigFile, "-Mkoalabear="+zigStatics)
			cmd.Dir = repoRoot
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "zig test failed:\n%s", out)
		})
	}
	if zigErr != nil {
		t.Skip("zig binary is not installed; generated Zig files were written but not compiled")
	}
}

func findGlobalVerifier(t *testing.T, sys *wiop.System) *global.Verifier {
	t.Helper()
	for _, r := range sys.Rounds {
		for _, action := range r.VerifierActions {
			if verifier, ok := action.(*global.Verifier); ok {
				return verifier
			}
		}
	}
	require.FailNow(t, "compiled system has no global verifier action")
	return nil
}

func repoRootFromTest(t *testing.T) string {
	t.Helper()
	_, file, _, ok := goruntime.Caller(0)
	require.True(t, ok, "runtime.Caller should locate the test file")
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func zigVerifierFileName(scenarioName string) string {
	return "global_verifier_" + strings.ToLower(scenarioName) + ".zig"
}
