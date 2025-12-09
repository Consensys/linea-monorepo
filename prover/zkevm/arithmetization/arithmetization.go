package arithmetization

import (
	"errors"
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/asm"
	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/air"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
	"github.com/consensys/go-corset/pkg/util/field/koalabear"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// Settings specifies the parameters for the arithmetization part of the zkEVM.
type Settings struct {
	Limits *config.TracesLimits
	// IgnoreCompatibilityCheck disables the strong compatibility check.
	// Specifically, it does not require the constraints and the trace file to
	// have both originated from the same commit.  By default, the compability
	// check should be enabled.
	IgnoreCompatibilityCheck *bool
	// OptimisationLevel determines the optimisation level which go-corset will
	// apply when compiling the zkevm.bin file to AIR constraints.  If in doubt,
	// use mir.DEFAULT_OPTIMISATION_LEVEL.
	OptimisationLevel *mir.OptimisationConfig
}

// Arithmetization exposes all the methods relevant for the user to interact
// with the arithmetization of the zkEVM. It is a sub-component of the whole
// ZkEvm object as it does not includes the precompiles, the keccaks and the
// signature verification.
type Arithmetization struct {
	Settings *Settings
	// ZkEVMBin contains the zkevm.bin file as a byte array. It is kept in the
	// struct as it is used for serialization.
	ZkEVMBin []byte
	// Binary encoding of the zkevm.bin file, which captures the high-level
	// structure of constraints.  This is primarily useful for assembly
	// functions (as these have big differences between their assembly
	// representation and their constraints representation).
	BinaryFile *binfile.BinaryFile `serde:"omit"`
	// Air schema defines the low-level columns, constraints and computations
	// used to expand a given trace, and to subsequently to check
	// satisfiability.
	AirSchema *air.Schema[koalabear.Element] `serde:"omit"`
	// Maps each column in the raw trace file into one (or more) columns in the
	// expanded trace file.  In particular, columns which are too large for the
	// given field are split into multiple "limbs".
	LimbMapping module.LimbsMap `serde:"omit"`
	// Metadata embedded in the zkevm.bin file, as needed to check
	// compatibility.  Guaranteed non-nil.
	Metadata typed.Map `serde:"omit"`
}

// NewArithmetization is the function that declares all the columns and the constraints of
// the zkEVM in the input builder object.
func NewArithmetization(builder *wizard.Builder, settings Settings) *Arithmetization {
	// Read and parse the binary file
	binf, metadata, errS := ReadZkevmBin()
	if errS != nil {
		panic(errS)
	}
	// Compile binary file into an air.Schema
	schema, mapping := CompileZkevmBin(binf, settings.OptimisationLevel)
	// Translate air.Schema into prover's internal representation
	Define(builder.CompiledIOP, schema, settings.Limits)
	// Done
	return &Arithmetization{
		BinaryFile:  binf,
		AirSchema:   schema,
		Settings:    &settings,
		LimbMapping: mapping,
		Metadata:    metadata,
		ZkEVMBin:    []byte(zkevmStr),
	}
}

// Assign the arithmetization related columns. Namely, it will open the file
// specified in the witness object, call corset and assign the prover runtime
// columns. As part of the assignment processs, the original trace is expanded
// according to the given schema.  The expansion process is about filling in
// computed columns with concrete values, such for determining multiplicative
// inverses, etc.
func (a *Arithmetization) Assign(run *wizard.ProverRuntime, traceFile string) {
	var (
		errs []error
		//
		traceF = files.MustRead(traceFile)
		// Parse trace file and extract raw column data.
		rawTrace, metadata, errT = ReadLtTraces(traceF)
	)

	// Performs a compatibility check by comparing the constraints
	// commit of zkevm.bin with the constraints commit of the trace file.
	// Panics if an incompatibility is detected.
	if !*a.Settings.IgnoreCompatibilityCheck {
		var errors []string

		zkevmBinCommit, ok := a.Metadata.String("commit")
		if !ok {
			errors = append(errors, "missing constraints commit metadata in 'zkevm.bin'")
		}

		traceFileCommit, ok := metadata.String("commit")
		if !ok {
			errors = append(errors, "missing constraints commit metadata in '.lt' file")
		}

		// Check commit mismatch
		if zkevmBinCommit != traceFileCommit {
			errors = append(errors, fmt.Sprintf(
				"zkevm.bin incompatible with trace file (commit %s vs %s)",
				zkevmBinCommit, traceFileCommit,
			))
		}

		// Panic only if there are errors
		if len(errors) > 0 {
			logrus.Panic("compatibility check failed with error message:\n" + strings.Join(errors, "\n"))
		}
	} else {
		logrus.Info("Skip constraints compatibility check between zkevm.bin and trace file")
	}

	if errT != nil {
		fmt.Printf("error loading the trace fpath=%q err=%v", traceFile, errT.Error())
	}
	// Perform trace propagation
	rawTrace, errs = asm.Propagate(a.BinaryFile.Schema, rawTrace, true)
	// error check
	if len(errs) > 0 {
		logrus.Warnf("corset propagation gave the following errors: %v", errors.Join(errs...).Error())
	}
	// Perform trace expansion
	expandedTrace, errs := ir.NewTraceBuilder[koalabear.Element]().
		WithBatchSize(1024).
		WithRegisterMapping(a.LimbMapping).
		Build(a.AirSchema, rawTrace)
	//
	if len(errs) > 0 {
		logrus.Warnf("corset expansion gave the following errors: %v", errors.Join(errs...).Error())
	}
	// Passed
	AssignFromLtTraces(run, a.AirSchema, expandedTrace, a.Settings.Limits)
}

// LimbColumnsOf returns the wizard columns corresponding to the limbs for the
// tuple (moduleName, regName). The function furthermore ensures that the
// function has the requested number of limbs.
func (a *Arithmetization) LimbColumnsOf(comp *wizard.CompiledIOP, mod string, regName string, nLimbs int) []ifaces.Column {
	names := a.LimbsOf(mod, regName, nLimbs)
	cols := make([]ifaces.Column, len(names))
	for i, name := range names {
		cols[i] = comp.Columns.GetHandle(ifaces.ColID(name))
	}
	return cols
}

// LimbColumnsOfArr2 is sugar for
//
//	```
//	 	c := a.LimbColumnsOf(comp, mod, regName, 2)
//		return [2]ifaces.Column(c)
//	 ```
func (a *Arithmetization) LimbColumnsOfArr2(comp *wizard.CompiledIOP, mod string, regName string) [2]ifaces.Column {
	c := a.LimbColumnsOf(comp, mod, regName, 2)
	return [2]ifaces.Column(c)
}

// LimbColumnsOfArr3 is sugar for
//
//	```
//	 	c := a.LimbColumnsOf(comp, mod, regName, 3)
//		return [3]ifaces.Column(c)
//	 ```
func (a *Arithmetization) LimbColumnsOfArr3(comp *wizard.CompiledIOP, mod string, regName string) [3]ifaces.Column {
	c := a.LimbColumnsOf(comp, mod, regName, 3)
	return [3]ifaces.Column(c)
}

// LimbColumnsOfArr4 is sugar for
//
//	```
//	 	c := a.LimbColumnsOf(comp, mod, regName, 4)
//		return [4]ifaces.Column(c)
//	 ```
func (a *Arithmetization) LimbColumnsOfArr4(comp *wizard.CompiledIOP, mod string, regName string) [4]ifaces.Column {
	c := a.LimbColumnsOf(comp, mod, regName, 4)
	return [4]ifaces.Column(c)
}

// LimbColumnsOfArr8 is sugar for
//
//	```
//	 	c := a.LimbColumnsOf(comp, mod, regName, 8)
//		return [8]ifaces.Column(c)
//	 ```
func (a *Arithmetization) LimbColumnsOfArr8(comp *wizard.CompiledIOP, mod string, regName string) [8]ifaces.Column {
	c := a.LimbColumnsOf(comp, mod, regName, 8)
	return [8]ifaces.Column(c)
}

// LimbColumnsOfArr16 is sugar for
//
//	```
//	 	c := a.LimbColumnsOf(comp, mod, regName, 16)
//		return [16]ifaces.Column(c)
//	 ```
func (a *Arithmetization) LimbColumnsOfArr16(comp *wizard.CompiledIOP, mod string, regName string) [16]ifaces.Column {
	c := a.LimbColumnsOf(comp, mod, regName, 16)
	return [16]ifaces.Column(c)
}

// ColumnOf returns the wizard column associated with the given name and
// register. The function will fail with panic if the column is not found or the
// column has more than 1 register.
func (a *Arithmetization) ColumnOf(comp *wizard.CompiledIOP, name string, regName string) ifaces.Column {
	cols := a.LimbColumnsOf(comp, name, regName, 1)
	return cols[0]
}

// LimbsOf returns the fully qualified names of the limbs for the given abstract
// register in the given module.  The limbs are returned in little endian order
// (i.e. the least significant limb is at index 0).
func (a *Arithmetization) LimbsOf(mod string, regName string, nLimbs int) []string {
	// Identify limbs mapping for ecdata module
	var (
		// Extract the limb mapping for the given module
		modMap  = a.LimbMapping.ModuleOf(module.NewName(mod, 1))
		reg, ok = modMap.HasRegister(regName)
	)
	//
	if !ok {
		panic("malformed register limbs map")
	}
	// Extract limbs for given register (least significant limb comes first)
	limbs := modMap.LimbIds(reg)
	names := make([]string, len(limbs))
	// Sanity check we got the number of limbs we expected
	if len(limbs) != nLimbs {
		panic(fmt.Sprintf("incorrect number of limbs (expected %d found %d)", nLimbs, len(limbs)))
	}
	//
	for i, lid := range limbs {
		limb := modMap.Limb(lid)
		names[i] = fmt.Sprintf("%s.%s", modMap.Name(), limb.Name)
	}
	//
	return names
}
