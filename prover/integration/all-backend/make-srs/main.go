package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/constraint/solver"
)

var flagLagrange = flag.String("lagrange", "", "12345_bls12377, 123_bn254")

func main() {

	var t test_utils.FakeTestingT

	flagConfig := "integration/all-backend/config-integration-light.toml"
	var flagLagrange []string
	argv := os.Args[1:]
	for len(argv) != 0 {
		switch argv[0] {
		case "-c", "-config":
			flagConfig = argv[1]
			argv = argv[2:]
		case "-l", "-lagrange":
			argv = argv[1:]
			n := 0
			for n < len(argv) && argv[n][0] != '-' {
				n++
			}
			flagLagrange = argv[:n]
			argv = argv[n:]
		default:
			logrus.Errorf("unknown parameter '%s'", argv[0])
			logrus.Errorf("usage ./makesrs")
			t.Errorf("unknown parameter '%s'", argv[0])
			t.FailNow()
		}
	}

	flag.Parse()

	rootDir, err := blob.GetRepoRootPath()
	require.NoError(t, err)
	rootDir = filepath.Join(rootDir, "prover")
	logrus.Infof("switching working directory to '%s'", rootDir)
	require.NoError(t, os.Chdir(rootDir))

	cfg, err := config.NewConfigFromFile(flagConfig)
	require.NoError(t, err, "could not load config")

	store, err := circuits.NewSRSStore(cfg.PathForSRS())
	if err != nil {
		logrus.Errorf("creating SRS store: %v", err)
	}

	require.NoError(t, os.MkdirAll(cfg.PathForSRS(), 0600))
	var wg sync.WaitGroup
	createSRS := func(settings srsSpec) {
		logrus.Infof("checking for %s srs", settings.id.String())
		canonicalSize, _ := plonk.SRSSize(&settings)

		// check if it exists
		_, _, err = store.GetSRS(context.TODO(), &srsSpec{
			id:      settings.id,
			maxSize: settings.maxSize,
		})
		if err == nil {
			logrus.Infof("SRS found for %d-%s. Skipping creation", canonicalSize, settings.id.String())
			return
		}
		logrus.Errorln(err)
		logrus.Infof("could not load %s SRS. Creating instead.", settings.id.String())

		// it doesn't exist. create it.
		canonical, _, err := unsafekzg.NewSRS(&settings /*, unsafekzg.WithFSCache()*/)
		require.NoError(t, err)
		f, err := os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_canonical_%d_%s_aleo.memdump", canonicalSize, strings.Replace(settings.id.String(), "_", "", 1))), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, canonical.WriteDump(f, settings.maxSize))
		f.Close()

		wg.Done()
	}

	if len(flagLagrange) == 0 {
		logrus.Info("no lagrange parameter. Setting up canonical SRS")
		createSRS(srsSpec{ecc.BLS12_377, 1 << 27}) // TODO @Tabaie remove

		wg.Add(3)
		go createSRS(srsSpec{ecc.BLS12_377, 1 << 27})
		createSRS(srsSpec{ecc.BN254, 1 << 26})
		createSRS(srsSpec{ecc.BW6_761, 1 << 26})
		wg.Wait()
		return
	}

	logrus.Info("Lagrange parameter set. Assuming canonical SRS is set.")

	nameToCurve := map[string]ecc.ID{
		"bls12377": ecc.BLS12_377,
		"bn254":    ecc.BN254,
		"bw6761":   ecc.BW6_761,
	}

	writeLagrange := func(param string) {
		split := strings.Split(param, "_")
		if len(split) != 2 {
			panic("bad format")
		}
		size, err := strconv.Atoi(split[0])
		require.NoError(t, err)

		var match ecc.ID
		for k, v := range nameToCurve {
			if strings.HasPrefix(k, split[1]) { // a match
				logrus.Infof("argument '%s' matches curve name '%s'", split[1], k)
				if match != ecc.UNKNOWN {
					t.Errorf("multiple matches")
					t.FailNow()
				}
				match = v
			}
		}
		require.NotEqual(t, ecc.UNKNOWN, match, "argument '%s' doesn't match any supported curves", split[1])

		_, lagrange, err := store.GetSRS(context.TODO(), &srsSpec{match, size})
		require.NoError(t, err)
		f, err := os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_lagrange_%d_%s_aleo.memdump", size, strings.ReplaceAll(match.String(), "_", ""))), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, lagrange.WriteDump(f, size))
		f.Close()
		wg.Done()
	}

	wg.Add(len(flagLagrange))
	for _, lagrange := range flagLagrange[1:] {
		go writeLagrange(lagrange)
	}
	writeLagrange(flagLagrange[0])
	wg.Wait()
}

type srsSpec struct {
	id      ecc.ID
	maxSize int
}

// Add implements constraint.ConstraintSystem.
func (s *srsSpec) Add(a constraint.Element, b constraint.Element) constraint.Element {
	panic("unimplemented")
}

// AddBlueprint implements constraint.ConstraintSystem.
func (s *srsSpec) AddBlueprint(b constraint.Blueprint) constraint.BlueprintID {
	panic("unimplemented")
}

// AddCoeff implements constraint.ConstraintSystem.
func (s *srsSpec) AddCoeff(coeff constraint.Element) uint32 {
	panic("unimplemented")
}

// AddCommitment implements constraint.ConstraintSystem.
func (s *srsSpec) AddCommitment(c constraint.Commitment) error {
	panic("unimplemented")
}

// AddGkr implements constraint.ConstraintSystem.
func (s *srsSpec) AddGkr(gkr constraint.GkrInfo) error {
	panic("unimplemented")
}

// AddInstruction implements constraint.ConstraintSystem.
func (s *srsSpec) AddInstruction(bID constraint.BlueprintID, calldata []uint32) []uint32 {
	panic("unimplemented")
}

// AddInternalVariable implements constraint.ConstraintSystem.
func (s *srsSpec) AddInternalVariable() int {
	panic("unimplemented")
}

// AddLog implements constraint.ConstraintSystem.
func (s *srsSpec) AddLog(l constraint.LogEntry) {
	panic("unimplemented")
}

// AddPublicVariable implements constraint.ConstraintSystem.
func (s *srsSpec) AddPublicVariable(name string) int {
	panic("unimplemented")
}

// AddSecretVariable implements constraint.ConstraintSystem.
func (s *srsSpec) AddSecretVariable(name string) int {
	panic("unimplemented")
}

// AddSolverHint implements constraint.ConstraintSystem.
func (s *srsSpec) AddSolverHint(f solver.Hint, id solver.HintID, input []constraint.LinearExpression, nbOutput int) (internalVariables []int, err error) {
	panic("unimplemented")
}

// AttachDebugInfo implements constraint.ConstraintSystem.
func (s *srsSpec) AttachDebugInfo(debugInfo constraint.DebugInfo, constraintID []int) {
	panic("unimplemented")
}

// CheckUnconstrainedWires implements constraint.ConstraintSystem.
func (s *srsSpec) CheckUnconstrainedWires() error {
	panic("unimplemented")
}

// CoeffToString implements constraint.ConstraintSystem.
func (s *srsSpec) CoeffToString(coeffID int) string {
	panic("unimplemented")
}

// Field implements constraint.ConstraintSystem.
func (s *srsSpec) Field() *big.Int {
	return s.id.ScalarField()
}

// FieldBitLen implements constraint.ConstraintSystem.
func (s *srsSpec) FieldBitLen() int {
	panic("unimplemented")
}

// FromInterface implements constraint.ConstraintSystem.
func (s *srsSpec) FromInterface(interface{}) constraint.Element {
	panic("unimplemented")
}

// GetCoefficient implements constraint.ConstraintSystem.
func (s *srsSpec) GetCoefficient(i int) constraint.Element {
	panic("unimplemented")
}

// GetCommitments implements constraint.ConstraintSystem.
func (s *srsSpec) GetCommitments() constraint.Commitments {
	panic("unimplemented")
}

// GetInstruction implements constraint.ConstraintSystem.
func (s *srsSpec) GetInstruction(int) constraint.Instruction {
	panic("unimplemented")
}

// GetNbCoefficients implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbCoefficients() int {
	panic("unimplemented")
}

// GetNbConstraints implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbConstraints() int {
	return s.maxSize
}

// GetNbInstructions implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbInstructions() int {
	panic("unimplemented")
}

// GetNbInternalVariables implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbInternalVariables() int {
	panic("unimplemented")
}

// GetNbPublicVariables implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbPublicVariables() int {
	return 0
}

// GetNbSecretVariables implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbSecretVariables() int {
	panic("unimplemented")
}

// GetNbVariables implements constraint.ConstraintSystem.
func (s *srsSpec) GetNbVariables() (internal int, secret int, public int) {
	panic("unimplemented")
}

// Inverse implements constraint.ConstraintSystem.
func (s *srsSpec) Inverse(a constraint.Element) (constraint.Element, bool) {
	panic("unimplemented")
}

// IsOne implements constraint.ConstraintSystem.
func (s *srsSpec) IsOne(constraint.Element) bool {
	panic("unimplemented")
}

// IsSolved implements constraint.ConstraintSystem.
func (s *srsSpec) IsSolved(witness witness.Witness, opts ...solver.Option) error {
	panic("unimplemented")
}

// MakeTerm implements constraint.ConstraintSystem.
func (s *srsSpec) MakeTerm(coeff constraint.Element, variableID int) constraint.Term {
	panic("unimplemented")
}

// Mul implements constraint.ConstraintSystem.
func (s *srsSpec) Mul(a constraint.Element, b constraint.Element) constraint.Element {
	panic("unimplemented")
}

// Neg implements constraint.ConstraintSystem.
func (s *srsSpec) Neg(a constraint.Element) constraint.Element {
	panic("unimplemented")
}

// NewDebugInfo implements constraint.ConstraintSystem.
func (s *srsSpec) NewDebugInfo(errName string, i ...interface{}) constraint.DebugInfo {
	panic("unimplemented")
}

// One implements constraint.ConstraintSystem.
func (s *srsSpec) One() constraint.Element {
	panic("unimplemented")
}

// ReadFrom implements constraint.ConstraintSystem.
func (s *srsSpec) ReadFrom(r io.Reader) (n int64, err error) {
	panic("unimplemented")
}

// Solve implements constraint.ConstraintSystem.
func (s *srsSpec) Solve(witness witness.Witness, opts ...solver.Option) (any, error) {
	panic("unimplemented")
}

// String implements constraint.ConstraintSystem.
func (s *srsSpec) String(constraint.Element) string {
	panic("unimplemented")
}

// Sub implements constraint.ConstraintSystem.
func (s *srsSpec) Sub(a constraint.Element, b constraint.Element) constraint.Element {
	panic("unimplemented")
}

// ToBigInt implements constraint.ConstraintSystem.
func (s *srsSpec) ToBigInt(constraint.Element) *big.Int {
	panic("unimplemented")
}

// Uint64 implements constraint.ConstraintSystem.
func (s *srsSpec) Uint64(constraint.Element) (uint64, bool) {
	panic("unimplemented")
}

// VariableToString implements constraint.ConstraintSystem.
func (s *srsSpec) VariableToString(variableID int) string {
	panic("unimplemented")
}

// WriteTo implements constraint.ConstraintSystem.
func (s *srsSpec) WriteTo(w io.Writer) (n int64, err error) {
	panic("unimplemented")
}
