// codehashconsistency implements the necessary constraints to enforce consistency
// between the mimccodehash module and the statesummary module. The constraints
// generated in this package essentially aim at ensuring that
package codehashconsistency

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/mimccodehash"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
)

// Module stores the column expressing the consistency between the
// [mimccodehash.Module] (computing the MiMCCodeHash from the code in the ROM
// module and exposing the corresponding codehash) and the [statesummary.Module]
// (recording the access to the state, thus holding the KeccakCodeHash and the
// MiMCCode hash of the registered contracts). The consistency check is achieved
// by ensuring that if a Keccak is present in both the MiMCCodeHash and the
// StateSummary, then the corresponding KeccakCodeHash must be equal. This
// ensures that code exposed in the ROM module is the same as what is in the
// state.
type Module struct {
	StateSummaryInput *statesummary.Module
	MimcCodeHashInput *mimccodehash.Module

	IsActive       ifaces.Column
	StateSumKeccak common.HiLoColumns
	StateSumMiMC   ifaces.Column
	RomKeccak      common.HiLoColumns
	RomMiMC        ifaces.Column

	StateSumKeccakLimbs byte32cmp.LimbColumns
	RomKeccakLimbs      byte32cmp.LimbColumns

	CptStateSumKeccakLimbsHi wizard.ProverAction
	CptStateSumKeccakLimbsLo wizard.ProverAction
	CptRomKeccakLimbsHi      wizard.ProverAction
	CptRomKeccakLimbsLo      wizard.ProverAction
	CmpStateSumLimbs         wizard.ProverAction
	CmpRomLimbs              wizard.ProverAction
	ComRomVsStateSumLimbs    wizard.ProverAction

	StateSumIsGtRom, StateSumIsEqROM, StateSumIsLtRom ifaces.Column
	StateSumIsConst, StateSumIncreased                ifaces.Column
	RomIsConst, RomIncreased                          ifaces.Column
}

// NewModule returns a constrained [Module] connecting `ss` with `mch`. `name`
// is used as a prefix for the name of all the generated columns and constraints.
func NewModule(comp *wizard.CompiledIOP, name string, ss *statesummary.Module, mch *mimccodehash.Module) Module {

	name = name + "_CODEHASH_CONSISTENCY"
	size := max(ss.IsActive.Size(), mch.IsActive.Size())

	ch := Module{
		StateSumKeccak: common.NewHiLoColumns(comp, size, name+"_STATE_SUMMARY_KECCAK"),
		RomKeccak:      common.NewHiLoColumns(comp, size, name+"_ROM_KECCAK"),
		StateSumMiMC:   comp.InsertCommit(0, ifaces.ColID(name+"_STATE_SUMMARY_MIMC"), size),
		RomMiMC:        comp.InsertCommit(0, ifaces.ColID(name+"_STATE_SUMMARY_MIMC"), size),
	}

	var (
		romDecreased                     ifaces.Column
		stateSumDecreased                ifaces.Column
		romLimbsHi, romLimbsLo           byte32cmp.LimbColumns
		stateSumLimbsHi, stateSumLimbsLo byte32cmp.LimbColumns
	)

	romLimbsHi, ch.CptRomKeccakLimbsHi = byte32cmp.Decompose(comp, ch.RomKeccak.Hi, 8, 16)
	romLimbsLo, ch.CptRomKeccakLimbsLo = byte32cmp.Decompose(comp, ch.RomKeccak.Lo, 8, 16)
	stateSumLimbsHi, ch.CptStateSumKeccakLimbsHi = byte32cmp.Decompose(comp, ch.StateSumKeccak.Hi, 8, 16)
	stateSumLimbsLo, ch.CptStateSumKeccakLimbsLo = byte32cmp.Decompose(comp, ch.StateSumKeccak.Lo, 8, 16)
	ch.RomKeccakLimbs = byte32cmp.FuseLimbs(romLimbsHi, romLimbsLo)
	ch.StateSumKeccakLimbs = byte32cmp.FuseLimbs(stateSumLimbsHi, stateSumLimbsLo)

	ch.RomIncreased, ch.RomIsConst, romDecreased, ch.CmpRomLimbs = byte32cmp.CmpMultiLimbs(
		comp,
		ch.RomKeccakLimbs,
		ch.RomKeccakLimbs.Shift(-1),
	)

	ch.StateSumIncreased, ch.StateSumIsConst, stateSumDecreased, ch.CmpStateSumLimbs = byte32cmp.CmpMultiLimbs(
		comp,
		ch.StateSumKeccakLimbs,
		ch.StateSumKeccakLimbs.Shift(-1),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_ROM_IS_SORTED"),
		symbolic.NewVariable(romDecreased),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_STATE_SUMMARY_IS_SORTED"),
		symbolic.NewVariable(stateSumDecreased),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_STATE_SUM_STAY_SAME"),
		symbolic.Mul(
			ch.IsActive,
			symbolic.Sub(
				column.Shift(ch.StateSumIsLtRom, -1),
				ch.StateSumIsConst,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_ROM_STAY_SAME"),
		symbolic.Mul(
			ch.IsActive,
			symbolic.Sub(
				column.Shift(ch.StateSumIsGtRom, -1),
				ch.RomIsConst,
			),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_KECCAK_CONSISTENCY_HI"),
		symbolic.Mul(
			ch.StateSumIsEqROM,
			symbolic.Sub(ch.RomKeccak.Hi, ch.StateSumKeccak.Hi),
		),
	)

	comp.InsertGlobal(
		0,
		ifaces.QueryID(name+"_KECCAK_CONSISTENCY_LO"),
		symbolic.Mul(
			ch.StateSumIsEqROM,
			symbolic.Sub(ch.RomKeccak.Lo, ch.StateSumKeccak.Lo),
		),
	)

	comp.GenericFragmentedConditionalInclusion(
		0,
		ifaces.QueryID(name+"_IMPORT_STATE_SUMMARY_BACK"),
		[][]ifaces.Column{
			{
				ss.Account.Initial.MiMCCodeHash,
				ss.Account.Initial.KeccakCodeHash.Hi,
				ss.Account.Initial.KeccakCodeHash.Lo,
			},
			{
				ss.Account.Final.MiMCCodeHash,
				ss.Account.Final.KeccakCodeHash.Hi,
				ss.Account.Final.KeccakCodeHash.Lo,
			},
		},
		[]ifaces.Column{
			ch.StateSumMiMC,
			ch.StateSumKeccak.Hi,
			ch.StateSumKeccak.Lo,
		},
		[]ifaces.Column{
			ss.IsActive,
			ss.IsActive,
		},
		ch.IsActive,
	)

	comp.InsertInclusionDoubleConditional(
		0,
		ifaces.QueryIDf(name+"_IMPORT_STATE_SUMMARY_FORTH_INITIAL"),
		[]ifaces.Column{
			ch.StateSumMiMC,
			ch.StateSumKeccak.Hi,
			ch.StateSumKeccak.Lo,
		},
		[]ifaces.Column{
			ss.Account.Initial.MiMCCodeHash,
			ss.Account.Initial.KeccakCodeHash.Hi,
			ss.Account.Initial.KeccakCodeHash.Lo,
		},
		ch.IsActive,
		ss.IsActive,
	)

	comp.InsertInclusionDoubleConditional(
		0,
		ifaces.QueryIDf(name+"_IMPORT_STATE_SUMMARY_FORTH_FINAL"),
		[]ifaces.Column{
			ch.StateSumMiMC,
			ch.StateSumKeccak.Hi,
			ch.StateSumKeccak.Lo,
		},
		[]ifaces.Column{
			ss.Account.Final.MiMCCodeHash,
			ss.Account.Final.KeccakCodeHash.Hi,
			ss.Account.Final.KeccakCodeHash.Lo,
		},
		ss.IsActive,
		ch.IsActive,
	)

	comp.InsertInclusionConditionalOnIncluded(
		0,
		ifaces.QueryIDf(name+"_IMPORT_MIMC_CODE_HASH_FORTH"),
		[]ifaces.Column{
			ch.RomMiMC,
			ch.RomKeccak.Hi,
			ch.RomKeccak.Lo,
		},
		[]ifaces.Column{
			mch.NewState,
			mch.CodeHashHi,
			mch.CodeHashLo,
		},
		mch.IsHashEnd,
	)

	comp.InsertInclusionConditionalOnIncluding(
		0,
		ifaces.QueryIDf(name+"_IMPORT_MIMC_CODE_HASH_BACK"),
		[]ifaces.Column{
			mch.NewState,
			mch.CodeHashHi,
			mch.CodeHashLo,
		},
		[]ifaces.Column{
			ch.RomMiMC,
			ch.RomKeccak.Hi,
			ch.RomKeccak.Lo,
		},
		mch.IsHashEnd,
	)

	return ch
}
