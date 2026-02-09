package verifiercol

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// compile check to enforce the struct to belong to the corresponding interface
var _ VerifierCol = ExpandedProofOrVerifyingKeyColWithZero{}

type ExpandedProofOrVerifyingKeyColWithZero struct {
	Col       ifaces.Column
	Expansion int
}

// Round returns the round ID of the column and implements the [ifaces.Column]
// interface.
func (ex ExpandedProofOrVerifyingKeyColWithZero) Round() int {
	return ex.Col.Round()
}

// GetColID returns the column ID
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColID() ifaces.ColID {
	return ifaces.ColIDf("EXPANDED_%v_%v_WITH_ZERO", ex.Col.GetColID(), ex.Expansion)
}

// MustExists implements the [ifaces.Column] interface and always returns true.
func (ex ExpandedProofOrVerifyingKeyColWithZero) MustExists() {
	ex.Col.MustExists()
}

// Size returns the size of the colum and implements the [ifaces.Column]
// interface.
func (ex ExpandedProofOrVerifyingKeyColWithZero) Size() int {
	return ex.Col.Size() * ex.Expansion
}

// IsBase returns the base status of the column and implements the [ifaces.Column]
func (ex ExpandedProofOrVerifyingKeyColWithZero) IsBase() bool {
	return ex.Col.IsBase()
}

// GetColAssignment returns the assignment of the current column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	assi := ex.Col.GetColAssignment(run)
	if smartvectors.IsBase(assi) {
		values := make([][]field.Element, ex.Expansion)
		values[0] = smartvectors.IntoRegVec(assi)
		// zeros denote slice of zeros of length of the verifier column
		zeros := make([]field.Element, ex.Col.Size())
		// we start at 1 because the first slice is the actual verifier column
		// the other slices are zeros
		for j := 1; j < ex.Expansion; j++ {
			values[j] = zeros
		}
		res := vector.Interleave(values...)
		return smartvectors.NewRegular(res)
	}

	// extension case
	assiExt := smartvectors.IntoRegVecExt(assi)
	// prepare slices of extension elements
	valuesExt := make([][]fext.Element, ex.Expansion)
	valuesExt[0] = assiExt
	zerosExt := make([]fext.Element, ex.Col.Size())
	for j := 1; j < ex.Expansion; j++ {
		valuesExt[j] = zerosExt
	}
	// interleave into a single slice
	newSize := ex.Expansion * ex.Col.Size()
	resExt := make([]fext.Element, newSize)
	for i := 0; i < ex.Expansion; i++ {
		for j := 0; j < ex.Col.Size(); j++ {
			resExt[i+j*ex.Expansion] = valuesExt[i][j]
		}
	}
	return smartvectors.NewRegularExt(resExt)
}

// GetColAssignmentGnark returns a gnark assignment of the current column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnark(api frontend.API, run ifaces.GnarkRuntime) []koalagnark.Element {
	koalaAPI := koalagnark.NewAPI(api)

	assi := ex.Col.GetColAssignmentGnark(api, run)
	res := make([]koalagnark.Element, ex.Size())
	for i := 0; i < len(assi); i++ {
		res[i*ex.Expansion] = assi[i]
		for j := 1; j < ex.Expansion; j++ {
			res[j+i*ex.Expansion] = koalaAPI.Zero()
		}
	}
	return res
}

// GetColAssignmentGnarkBase returns a gnark assignment of the current column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnarkBase(api frontend.API, run ifaces.GnarkRuntime) ([]koalagnark.Element, error) {
	koalaAPI := koalagnark.NewAPI(api)
	assi, err := ex.Col.GetColAssignmentGnarkBase(api, run)
	if err != nil {
		return nil, err
	}
	res := make([]koalagnark.Element, ex.Size())
	for i := 0; i < len(assi); i++ {
		res[i*ex.Expansion] = assi[i]
		for j := 1; j < ex.Expansion; j++ {
			res[j+i*ex.Expansion] = koalaAPI.Zero()
		}
	}
	return res, nil
}

// GetColAssignmentGnarkExt returns a gnark assignment of the current column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnarkExt(api frontend.API, run ifaces.GnarkRuntime) []koalagnark.Ext {
	koalaAPI := koalagnark.NewAPI(api)
	assi := ex.Col.GetColAssignmentGnarkExt(api, run)
	res := make([]koalagnark.Ext, ex.Size())
	zeroExt := koalaAPI.ExtFrom(fext.Zero())
	for i := 0; i < len(assi); i++ {
		res[i*ex.Expansion] = assi[i]
		for j := 1; j < ex.Expansion; j++ {
			res[j+i*ex.Expansion] = zeroExt
		}
	}
	return res
}

// GetColAssignmentGnarkExt returns a gnark assignment of the current column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnarkExtAsPtr(api frontend.API, run ifaces.GnarkRuntime) []*koalagnark.Ext {
	koalaAPI := koalagnark.NewAPI(api)
	assi := ex.Col.GetColAssignmentGnarkExt(api, run)
	res := make([]*koalagnark.Ext, ex.Size())
	zeroExt := koalaAPI.ExtFrom(fext.Zero())
	for i := 0; i < len(assi); i++ {
		res[i*ex.Expansion] = &assi[i]
		for j := 1; j < ex.Expansion; j++ {
			res[j+i*ex.Expansion] = &zeroExt
		}
	}
	return res
}

// GetColAssignmentAt returns a particular position of the column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	if pos%ex.Expansion == 0 {
		return ex.Col.GetColAssignmentAt(run, pos/ex.Expansion)
	}
	return field.Zero()
}

// GetColAssignmentAtBase returns a particular position of the column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	if pos%ex.Expansion == 0 {
		r, err := ex.Col.GetColAssignmentAtBase(run, pos/ex.Expansion)
		if err != nil {
			return field.Zero(), err
		}
		return r, nil
	}
	return field.Zero(), nil
}

// GetColAssignmentAtExt returns a particular position of the column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	if pos%ex.Expansion == 0 {
		return ex.Col.GetColAssignmentAtExt(run, pos/ex.Expansion)
	}
	return fext.Zero()
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnarkAt(api frontend.API, run ifaces.GnarkRuntime, pos int) koalagnark.Element {
	koalaAPI := koalagnark.NewAPI(api)
	if pos%ex.Expansion == 0 {
		return ex.Col.GetColAssignmentGnarkAt(api, run, pos/ex.Expansion)
	}
	return koalaAPI.Zero()
}

// GetColAssignmentAtBase returns a particular position of the column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnarkAtBase(api frontend.API, run ifaces.GnarkRuntime, pos int) (koalagnark.Element, error) {
	koalaAPI := koalagnark.NewAPI(api)
	if pos%ex.Expansion == 0 {
		r, err := ex.Col.GetColAssignmentGnarkAtBase(api, run, pos/ex.Expansion)
		if err != nil {
			return koalaAPI.Zero(), err
		}
		return r, nil
	}
	return koalaAPI.Zero(), nil
}

// GetColAssignmentAtExt returns a particular position of the column
func (ex ExpandedProofOrVerifyingKeyColWithZero) GetColAssignmentGnarkAtExt(api frontend.API, run ifaces.GnarkRuntime, pos int) koalagnark.Ext {
	koalaAPI := koalagnark.NewAPI(api)
	if pos%ex.Expansion == 0 {
		return ex.Col.GetColAssignmentGnarkAtExt(api, run, pos/ex.Expansion)
	}
	return koalaAPI.ExtFrom(fext.Zero())
}

// IsComposite implements the [ifaces.Column] interface
func (ex ExpandedProofOrVerifyingKeyColWithZero) IsComposite() bool {
	return ex.Col.IsComposite()
}

// String implements the [symbolic.Metadata] interface
func (ex ExpandedProofOrVerifyingKeyColWithZero) String() string {
	return string(ex.GetColID())
}

// Split implements the [VerifierCol] interface,
// it is no-op for proofcols
func (ex ExpandedProofOrVerifyingKeyColWithZero) Split(_ *wizard.CompiledIOP, from, to int) ifaces.Column {
	return ex.Col
}
