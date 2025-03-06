package experiment

import (
	"errors"
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// DistributedWizard represents a wizard protocol that has undergone a
// distributed compilation process.
type DistributedWizard struct {

	// ModuleNames is the list of the names of the modules that compose
	// the distributed protocol.
	ModuleNames []ModuleName

	// LPPs is the list of the LPP parts for every modules
	LPPs []*ModuleLPP

	// GLs is the list of the GL parts for every modules
	GLs []*ModuleGL

	// Bootstrapper is the original compiledIOP precompiled with a few
	// preparation steps.
	Bootstrapper *wizard.CompiledIOP

	// Disc is the [ModuleDiscoverer] used to delimitate the scope for
	// each module.
	Disc ModuleDiscoverer
}

// DistributeWizard returns a [DistributedWizard] from a [wizard.CompiledIOP]. It
// takes ownership of the input [wizard.CompiledIOP]. And uses disc to design
// the scope of each module.
func DistributeWizard(comp *wizard.CompiledIOP, disc ModuleDiscoverer) DistributedWizard {

	if err := auditInitialWizard(comp); err != nil {
		utils.Panic("improper initial wizard for distribution: %v", err)
	}

	distributedWizard := DistributedWizard{
		Bootstrapper: precompileInitialWizard(comp),
	}

	disc.Analyze(distributedWizard.Bootstrapper)
	distributedWizard.ModuleNames = disc.ModuleList()

	for _, moduleName := range distributedWizard.ModuleNames {

		moduleFilter := moduleFilter{
			Disc:   disc,
			Module: moduleName,
		}

		filteredModuleInputs := moduleFilter.FilterCompiledIOP(
			distributedWizard.Bootstrapper,
		)

		distributedWizard.LPPs = append(
			distributedWizard.LPPs,
			BuildModuleLPP(&filteredModuleInputs),
		)

		distributedWizard.GLs = append(
			distributedWizard.GLs,
			BuildModuleGL(&filteredModuleInputs),
		)
	}

	return distributedWizard
}

// CompileModules applies the compilation steps to each modules identically.
func (dist *DistributedWizard) CompileModules(compilers ...func(*wizard.CompiledIOP)) {
	for i := range dist.ModuleNames {
		for _, compile := range compilers {
			compile(dist.LPPs[i].Wiop)
			compile(dist.GLs[i].Wiop)
		}
	}
}

// precompileInitialWizard pre-compiles the initial wizard protocol by applying all the
// compilation steps needing to be run before the splitting phase. Its role is to
// ensure that all of the queries that cannot be processed by the splitting phase
// are removed from the compiled IOP.
func precompileInitialWizard(comp *wizard.CompiledIOP) *wizard.CompiledIOP {
	mimc.CompileMiMC(comp)
	// specialqueries.RangeProof(comp)
	// specialqueries.CompileFixedPermutations(comp)
	logderivativesum.LookupIntoLogDerivativeSum(comp)
	permutation.CompileIntoGdProduct(comp)
	horner.ProjectionToHorner(comp)
	return comp
}

// auditInitialWizard scans the initial compiled-IOP and checks if the provided
// wizard is compatible with the [DistributedWizard]. This includes:
//
//   - Absence of precomputed columns in the wizard (except as fixed lookup tables)
//   - Absence of [verifiercol.VerifierCol] (except [verifiercol.ConstCol]) in all
//     queries.
//   - Selectors for Inclusions or Projections must be either [column.Natural] or
//     [verifiercol.ConstCol]
func auditInitialWizard(comp *wizard.CompiledIOP) error {

	allPrecomputedColumns := comp.Precomputed.ListAllKeys()
	var err error

	for _, col := range allPrecomputedColumns {
		err = errors.Join(err, fmt.Errorf("found precomputed column: %v", col))
	}

	allGlobalQueries := comp.QueriesNoParams.AllKeys()
	for _, qname := range allGlobalQueries {

		q := comp.QueriesNoParams.Data(qname)

		switch q_ := q.(type) {

		case query.Inclusion:
			shfted := q_.GetShiftedSelector()
			if len(shfted) > 0 {
				err = errors.Join(err, fmt.Errorf("inclusion query %v with shifted selectors %v", qname, shfted))
			}

		case query.Projection:
			shfted := q_.GetShiftedSelector()
			if len(shfted) > 0 {
				err = errors.Join(err, fmt.Errorf("inclusion query %v with shifted selectors %v", qname, shfted))
			}
		}

		err = errors.Join(err)
	}

	return err
}
