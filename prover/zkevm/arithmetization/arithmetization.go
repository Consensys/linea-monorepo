package arithmetization

import (
	"errors"
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/air"
	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// Settings specifies the parameters for the arithmetization part of the zkEVM.
type Settings struct {
	Limits *config.TracesLimits
	IgnoreCompatibilityCheck *bool
	OptimisationLevel        *mir.OptimisationConfig
}

// Arithmetization exposes all the methods relevant for the user to interact
// with the arithmetization of the zkEVM.
type Arithmetization struct {
	Settings *Settings
	Schema   *air.Schema
	Metadata typed.Map
}

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

func (a *Arithmetization) Assign(run *wizard.ProverRuntime, traceFile string) {
	traceF := files.MustRead(traceFile)
	rawColumns, metadata, errT := ReadLtTraces(traceF, a.Schema)

	if !*a.Settings.IgnoreCompatibilityCheck {
		var compatibilityErrors []string

		zkevmBinCommit, ok := a.Metadata.String("commit")
		if !ok {
			compatibilityErrors = append(compatibilityErrors, "missing constraints commit metadata in 'zkevm.bin'")
		}

		traceFileCommit, ok := metadata.String("commit")
		if !ok {
			compatibilityErrors = append(compatibilityErrors, "missing constraints commit metadata in '.lt' file")
		}

		if zkevmBinCommit != traceFileCommit {
			compatibilityErrors = append(compatibilityErrors, fmt.Sprintf(
				"zkevm.bin incompatible with trace file (commit %s vs %s)",
				zkevmBinCommit, traceFileCommit,
			))
		}

		if len(compatibilityErrors) > 0 {
			logrus.Panic("compatibility check failed with error message:\n" + strings.Join(compatibilityErrors, "\n"))
		}
	} else {
		logrus.Info("Skip constraints compatibility check between zkevm.bin and trace file")
	}

	if errT != nil {
		fmt.Printf("error loading the trace fpath=%q err=%v", traceFile, errT.Error())
	}

	// 🛠️ Patch: safer handling of expansion failure
	expandedTrace, errs := schema.NewTraceBuilder(a.Schema).Build(rawColumns)
	if len(errs) > 0 || expandedTrace == nil {
		var errMsg string
		if len(errs) > 0 {
			errMsg = errors.Join(errs...).Error()
		}
		if expandedTrace == nil {
			if errMsg != "" {
				errMsg += "; "
			}
			errMsg += "corset expansion returned nil trace"
		}
		logrus.Panic("corset expansion failed: " + errMsg)
	}

	AssignFromLtTraces(run, a.Schema, expandedTrace, a.Settings.Limits)
}
