package arithmetization

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/ethereum/go-ethereum/common"
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

// SanityCheckOptions holds optional parameters for sanity checking
// to check consistency between the prover response vs trace metadata
type SanityCheckOptions struct {
	ChainID uint
	// NbAllL2L1MessageHashes stores the total number of L2 to L1 message hashes.
	L2BridgeAddress        common.Address
	NbAllL2L1MessageHashes int
	L2L1MessageLimits      int
}

// Arithmetization exposes all the methods relevant for the user to interact
// with the arithmetization of the zkEVM. It is a sub-component of the whole
// ZkEvm object as it does not includes the precompiles, the keccaks and the
// signature verification.
type Arithmetization struct {
	Settings *Settings
	// Schema defines the columns, constraints and computations used to expand a
	// given trace, and to subsequently to check satisfiability.
	Schema *air.Schema
	// Metadata embedded in the zkevm.bin file, as needed to check
	// compatibility.  Guaranteed non-nil.
	Metadata typed.Map
}

// NewArithmetization is the function that declares all the columns and the constraints of
// the zkEVM in the input builder object.
func NewArithmetization(builder *wizard.Builder, settings Settings) *Arithmetization {
	schema, metadata, errS := ReadZkevmBin(settings.OptimisationLevel)
	if errS != nil {
		panic(errS)
	}

	Define(builder.CompiledIOP, schema, settings.Limits)

	return &Arithmetization{
		Schema:   schema,
		Settings: &settings,
		Metadata: metadata,
	}
}

// Assign the arithmetization related columns. Namely, it will open the file
// specified in the witness object, call corset and assign the prover runtime
// columns. As part of the assignment processs, the original trace is expanded
// according to the given schema.  The expansion process is about filling in
// computed columns with concrete values, such for determining multiplicative
// inverses, etc.
func (a *Arithmetization) Assign(run *wizard.ProverRuntime, traceFile string, proverInput *SanityCheckOptions) {
	traceF := files.MustRead(traceFile)
	// Parse trace file and extract raw column data
	rawColumns, metadata, errT := ReadLtTraces(traceF, a.Schema)

	// Perform constraints compatibility check (trace file vs zkevm.bin)
	compatibilityCheck(metadata, a)

	// Perform sanity checks (trace file and prover response)
	if proverInput != nil {
		proverInputSanityCheck(metadata, proverInput)
	}

	if errT != nil {
		fmt.Printf("error loading the trace fpath=%q err=%v", traceFile, errT.Error())
	}

	// Perform trace expansion
	expandedTrace, errs := schema.NewTraceBuilder(a.Schema).Build(rawColumns)
	if len(errs) > 0 {
		logrus.Warnf("corset expansion gave the following errors: %v", errors.Join(errs...).Error())
	}
	// Passed
	AssignFromLtTraces(run, a.Schema, expandedTrace, a.Settings.Limits)
}

// compatibilityCheck ensures the constraints commit of zkevm.bin matches the trace metadata.
// It performs a compatibility check by comparing the constraints commit of zkevm.bin
// with the constraints commit of the trace file, panicking if an incompatibility is found.
func compatibilityCheck(metadata typed.Map, a *Arithmetization) {
	if *a.Settings.IgnoreCompatibilityCheck == false {
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
		} else {
			logrus.Info("zkevm.bin and trace file passed constraints compatibility check")
		}
	} else {
		logrus.Info("Skip constraints compatibility check between zkevm.bin and trace file")
	}
}

// sanityCheck performs sanity checks between the prover response and the trace file.
// It verifies that the chainID and total L2 to L1 message logs are consistent between
// the two sources, and panics if a mismatch is detected, before expanding the trace.
func proverInputSanityCheck(traceMetadata typed.Map, proverInput *SanityCheckOptions) {
	// Sanity-check: Chain ID
	// extract chainID from the .lt file
	traceChainIDStr, ok := traceMetadata.String("chainId")
	if !ok {
		logrus.Panic("chainId is missing or not an integer")
	}
	// Convert string to int
	traceChainID, err := strconv.Atoi(traceChainIDStr)
	if err != nil {
		logrus.Panicf("invalid chainId format: %s", traceChainIDStr)
	}
	// sanity-check if chainID matches
	if int(proverInput.ChainID) != traceChainID {
		logrus.Panicf("sanity-check failed: responseChainID=%v vs traceChainID=%v", int(proverInput.ChainID), traceChainID)
	}

	// l2l1 bridge address is same in both trace and prover response

	// Sanity-check: L2 to L1 Messages
	// extract lineCounts BLOCK_L2_L1_LOGS from the .lt file
	BLOCK_L2_L1_LOGS_Str, ok := traceMetadata.String("BLOCK_L2_L1_LOGS")
	if !ok {
		logrus.Panic("missing BLOCK_L2_L1_LOGS metadata in .lt file")
	}
	// Convert string to int
	BLOCK_L2_L1_LOGS, err := strconv.Atoi(BLOCK_L2_L1_LOGS_Str)
	if err != nil {
		logrus.Panicf("invalid BLOCK_L2_L1_LOGS format: %s", BLOCK_L2_L1_LOGS_Str)
	}
	// sanity-check if there is a mismatch between prover response and trace metadata
	if proverInput.NbAllL2L1MessageHashes != BLOCK_L2_L1_LOGS {
		logrus.Panicf("sanity-check failed: prover response NbAllL2L1MessageHashes=%v\n"+
			"vs trace lineCounts(BLOCK_L2_L1_LOGS)=%v",
			proverInput.NbAllL2L1MessageHashes, BLOCK_L2_L1_LOGS)
	}

	// range check:  linecounts L2 to L1 Messages <= target (limit)

}
