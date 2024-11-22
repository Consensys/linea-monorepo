package allbackend

import (
	"io"
	"math/big"
	"os"
	"path/filepath"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func CdProver(t require.TestingT) {
	rootDir, err := blob.GetRepoRootPath()
	require.NoError(t, err)
	rootDir = filepath.Join(rootDir, "prover")
	logrus.Infof("switching working directory to '%s'", rootDir)
	require.NoError(t, os.Chdir(rootDir))

}

type SrsSpec struct {
	Id      ecc.ID
	MaxSize int
}

// Add implements constraint.ConstraintSystem.
func (s *SrsSpec) Add(a constraint.Element, b constraint.Element) constraint.Element {
	panic("unimplemented")
}

// AddBlueprint implements constraint.ConstraintSystem.
func (s *SrsSpec) AddBlueprint(b constraint.Blueprint) constraint.BlueprintID {
	panic("unimplemented")
}

// AddCoeff implements constraint.ConstraintSystem.
func (s *SrsSpec) AddCoeff(coeff constraint.Element) uint32 {
	panic("unimplemented")
}

// AddCommitment implements constraint.ConstraintSystem.
func (s *SrsSpec) AddCommitment(c constraint.Commitment) error {
	panic("unimplemented")
}

// AddGkr implements constraint.ConstraintSystem.
func (s *SrsSpec) AddGkr(gkr constraint.GkrInfo) error {
	panic("unimplemented")
}

// AddInstruction implements constraint.ConstraintSystem.
func (s *SrsSpec) AddInstruction(bID constraint.BlueprintID, calldata []uint32) []uint32 {
	panic("unimplemented")
}

// AddInternalVariable implements constraint.ConstraintSystem.
func (s *SrsSpec) AddInternalVariable() int {
	panic("unimplemented")
}

// AddLog implements constraint.ConstraintSystem.
func (s *SrsSpec) AddLog(l constraint.LogEntry) {
	panic("unimplemented")
}

// AddPublicVariable implements constraint.ConstraintSystem.
func (s *SrsSpec) AddPublicVariable(name string) int {
	panic("unimplemented")
}

// AddSecretVariable implements constraint.ConstraintSystem.
func (s *SrsSpec) AddSecretVariable(name string) int {
	panic("unimplemented")
}

// AddSolverHint implements constraint.ConstraintSystem.
func (s *SrsSpec) AddSolverHint(f solver.Hint, id solver.HintID, input []constraint.LinearExpression, nbOutput int) (internalVariables []int, err error) {
	panic("unimplemented")
}

// AttachDebugInfo implements constraint.ConstraintSystem.
func (s *SrsSpec) AttachDebugInfo(debugInfo constraint.DebugInfo, constraintID []int) {
	panic("unimplemented")
}

// CheckUnconstrainedWires implements constraint.ConstraintSystem.
func (s *SrsSpec) CheckUnconstrainedWires() error {
	panic("unimplemented")
}

// CoeffToString implements constraint.ConstraintSystem.
func (s *SrsSpec) CoeffToString(coeffID int) string {
	panic("unimplemented")
}

// Field implements constraint.ConstraintSystem.
func (s *SrsSpec) Field() *big.Int {
	return s.Id.ScalarField()
}

// FieldBitLen implements constraint.ConstraintSystem.
func (s *SrsSpec) FieldBitLen() int {
	panic("unimplemented")
}

// FromInterface implements constraint.ConstraintSystem.
func (s *SrsSpec) FromInterface(interface{}) constraint.Element {
	panic("unimplemented")
}

// GetCoefficient implements constraint.ConstraintSystem.
func (s *SrsSpec) GetCoefficient(i int) constraint.Element {
	panic("unimplemented")
}

// GetCommitments implements constraint.ConstraintSystem.
func (s *SrsSpec) GetCommitments() constraint.Commitments {
	panic("unimplemented")
}

// GetInstruction implements constraint.ConstraintSystem.
func (s *SrsSpec) GetInstruction(int) constraint.Instruction {
	panic("unimplemented")
}

// GetNbCoefficients implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbCoefficients() int {
	panic("unimplemented")
}

// GetNbConstraints implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbConstraints() int {
	return s.MaxSize
}

// GetNbInstructions implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbInstructions() int {
	panic("unimplemented")
}

// GetNbInternalVariables implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbInternalVariables() int {
	panic("unimplemented")
}

// GetNbPublicVariables implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbPublicVariables() int {
	return 0
}

// GetNbSecretVariables implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbSecretVariables() int {
	panic("unimplemented")
}

// GetNbVariables implements constraint.ConstraintSystem.
func (s *SrsSpec) GetNbVariables() (internal int, secret int, public int) {
	panic("unimplemented")
}

// Inverse implements constraint.ConstraintSystem.
func (s *SrsSpec) Inverse(a constraint.Element) (constraint.Element, bool) {
	panic("unimplemented")
}

// IsOne implements constraint.ConstraintSystem.
func (s *SrsSpec) IsOne(constraint.Element) bool {
	panic("unimplemented")
}

// IsSolved implements constraint.ConstraintSystem.
func (s *SrsSpec) IsSolved(witness witness.Witness, opts ...solver.Option) error {
	panic("unimplemented")
}

// MakeTerm implements constraint.ConstraintSystem.
func (s *SrsSpec) MakeTerm(coeff constraint.Element, variableID int) constraint.Term {
	panic("unimplemented")
}

// Mul implements constraint.ConstraintSystem.
func (s *SrsSpec) Mul(a constraint.Element, b constraint.Element) constraint.Element {
	panic("unimplemented")
}

// Neg implements constraint.ConstraintSystem.
func (s *SrsSpec) Neg(a constraint.Element) constraint.Element {
	panic("unimplemented")
}

// NewDebugInfo implements constraint.ConstraintSystem.
func (s *SrsSpec) NewDebugInfo(errName string, i ...interface{}) constraint.DebugInfo {
	panic("unimplemented")
}

// One implements constraint.ConstraintSystem.
func (s *SrsSpec) One() constraint.Element {
	panic("unimplemented")
}

// ReadFrom implements constraint.ConstraintSystem.
func (s *SrsSpec) ReadFrom(r io.Reader) (n int64, err error) {
	panic("unimplemented")
}

// Solve implements constraint.ConstraintSystem.
func (s *SrsSpec) Solve(witness witness.Witness, opts ...solver.Option) (any, error) {
	panic("unimplemented")
}

// String implements constraint.ConstraintSystem.
func (s *SrsSpec) String(constraint.Element) string {
	panic("unimplemented")
}

// Sub implements constraint.ConstraintSystem.
func (s *SrsSpec) Sub(a constraint.Element, b constraint.Element) constraint.Element {
	panic("unimplemented")
}

// ToBigInt implements constraint.ConstraintSystem.
func (s *SrsSpec) ToBigInt(constraint.Element) *big.Int {
	panic("unimplemented")
}

// Uint64 implements constraint.ConstraintSystem.
func (s *SrsSpec) Uint64(constraint.Element) (uint64, bool) {
	panic("unimplemented")
}

// VariableToString implements constraint.ConstraintSystem.
func (s *SrsSpec) VariableToString(variableID int) string {
	panic("unimplemented")
}

// WriteTo implements constraint.ConstraintSystem.
func (s *SrsSpec) WriteTo(w io.Writer) (n int64, err error) {
	panic("unimplemented")
}
