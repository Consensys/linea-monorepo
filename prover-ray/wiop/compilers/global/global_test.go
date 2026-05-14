package global_test

import (
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
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

func TestVerifier_CheckZig_Fibonacci(t *testing.T) {
	zigPath, err := exec.LookPath("zig")
	if err != nil {
		t.Skip("zig binary is not installed")
	}

	sc := wioptest.NewFibonacciVanishingScenario()
	global.Compile(sc.Sys)

	rt := wiop.NewRuntime(sc.Sys)
	sc.AssignHonest(&rt)
	require.NoError(t, wioptest.RunAndVerify(&rt),
		"Go verifier must accept before generating the equivalent Zig verifier")

	tmp := t.TempDir()
	outputDir := filepath.Join(tmp, "zigverifiers")
	zigStatics := copyZigStatics(t, tmp)
	cmdDir := tmp
	if os.Getenv("LINEA_KEEP_ZIG_VERIFIER") == "1" {
		repoRoot := repoRootFromTest(t)
		outputDir = filepath.Join(repoRoot, "wiop", "zigverifiers")
		zigStatics = filepath.Join(repoRoot, "wiop", "zigstatics", "koalabear_field.zig")
		cmdDir = repoRoot
	}

	zigFile, err := findGlobalVerifier(t, sc.Sys).CheckZigToDir(rt, outputDir)
	require.NoError(t, err)
	require.FileExists(t, zigFile)

	//nolint:gosec // Test executes the local Zig compiler on generated verifier source.
	cmd := exec.Command(zigPath, "test", "--dep", "koalabear", "-Mroot="+zigFile, "-Mkoalabear="+zigStatics)
	cmd.Env = append(os.Environ(),
		"ZIG_GLOBAL_CACHE_DIR="+filepath.Join(tmp, "global-cache"),
		"ZIG_LOCAL_CACHE_DIR="+filepath.Join(tmp, "local-cache"),
	)
	cmd.Dir = cmdDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "zig test failed:\n%s", out)
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

func copyZigStatics(t *testing.T, tmp string) string {
	t.Helper()
	src := filepath.Join(repoRootFromTest(t), "wiop", "zigstatics", "koalabear_field.zig")
	dstDir := filepath.Join(tmp, "zigstatics")
	require.NoError(t, os.MkdirAll(dstDir, 0o755))

	data, err := os.ReadFile(src)
	require.NoError(t, err)
	dst := filepath.Join(dstDir, "koalabear_field.zig")
	require.NoError(t, os.WriteFile(dst, data, 0o600))
	return dst
}

func repoRootFromTest(t *testing.T) string {
	t.Helper()
	_, file, _, ok := goruntime.Caller(0)
	require.True(t, ok, "runtime.Caller should locate the test file")
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
