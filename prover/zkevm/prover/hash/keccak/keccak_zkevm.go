// The keccak package accumulates the providers from different zkEVM modules,
//
//	and proves the hash consistency over the unified provider.
//
// The provider encodes the inputs and outputs of the hash from different modules.
package keccak

import (
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	gen_acc "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/acc_module"
)

type KeccakZkEVM struct {
	Settings *Settings

	// The [wizard.ProverAction] for submodules.
	pa_accData wizard.ProverAction
	pa_accInfo wizard.ProverAction
	pa_keccak  wizard.ProverAction
}

func NewKeccakZkEVM(comp *wizard.CompiledIOP, settings Settings, providersFromEcdsa []generic.GenericByteModule) *KeccakZkEVM {
	return newKeccakZkEvm(
		comp,
		settings, append(
			providersFromEcdsa,
			getShakiraArithmetization(comp),
			getRlpAddArithmetization(comp),
		),
	)
}

func newKeccakZkEvm(comp *wizard.CompiledIOP, settings Settings, providers []generic.GenericByteModule) *KeccakZkEVM {

	// create the list of  [generic.GenDataModule] and [generic.GenInfoModule]
	var (
		gdm = make([]generic.GenDataModule, 0, len(providers))
		gim = make([]generic.GenInfoModule, 0, len(providers))
	)

	for i := range providers {
		gdm = append(gdm, providers[i].Data)
		gim = append(gim, providers[i].Info)
	}

	var (
		inpAcc = gen_acc.GenericAccumulatorInputs{
			MaxNumKeccakF: settings.MaxNumKeccakf,
			ProvidersData: gdm,
			ProvidersInfo: gim,
		}

		// unify the data from different providers in a single provider
		accData = gen_acc.NewGenericDataAccumulator(comp, inpAcc)
		// unify the info from different providers in a single provider
		accInfo = gen_acc.NewGenericInfoAccumulator(comp, inpAcc)

		keccakInp = KeccakSingleProviderInput{
			Provider: generic.GenericByteModule{
				Data: accData.Provider,
				Info: accInfo.Provider,
			},
			MaxNumKeccakF: settings.MaxNumKeccakf,
		}
		keccak = NewKeccakSingleProvider(comp, keccakInp)
	)

	res := &KeccakZkEVM{
		pa_accData: accData,
		pa_accInfo: accInfo,
		pa_keccak:  keccak,
		Settings:   &settings,
	}
	return res

}

func (k *KeccakZkEVM) Run(run *wizard.ProverRuntime) {

	run.Spec.Columns.
		GetHandle("shakiradata.SELECTOR_KECCAK_RES_LO").(dedicated.ManuallyShifted).
		Assign(run)

	k.pa_accData.Run(run)
	k.pa_accInfo.Run(run)
	k.pa_keccak.Run(run)
}

func getShakiraArithmetization(comp *wizard.CompiledIOP) generic.GenericByteModule {

	res := generic.GenericByteModule{
		Data: generic.GenDataModule{
			HashNum: comp.Columns.GetHandle("shakiradata.ID"),
			Index:   comp.Columns.GetHandle("shakiradata.INDEX"),
			Limb:    comp.Columns.GetHandle("shakiradata.LIMB"),
			NBytes:  comp.Columns.GetHandle("shakiradata.nBYTES"),
			ToHash:  comp.Columns.GetHandle("shakiradata.IS_KECCAK_DATA"),
		},
		Info: generic.GenInfoModule{
			HashNum: comp.Columns.GetHandle("shakiradata.ID"),
			HashLo:  comp.Columns.GetHandle("shakiradata.LIMB"),
			HashHi:  comp.Columns.GetHandle("shakiradata.LIMB"),
			// Before, we usse to pass column.Shift(IsHashHi, -1) but this does
			// not work with the prover distribution as the column is used as
			// a filter for a projection query.
			IsHashHi: comp.Columns.GetHandle("shakiradata.SELECTOR_KECCAK_RES_HI"),
		},
	}

	res.Info.IsHashLo = dedicated.ManuallyShift(comp, res.Info.IsHashHi, -1)

	return res
}

func getRlpAddArithmetization(comp *wizard.CompiledIOP) generic.GenericByteModule {
	return generic.GenericByteModule{
		Data: generic.GenDataModule{
			HashNum: comp.Columns.GetHandle("rlpaddr.STAMP"),
			Index:   comp.Columns.GetHandle("rlpaddr.INDEX"),
			Limb:    comp.Columns.GetHandle("rlpaddr.LIMB"),
			NBytes:  comp.Columns.GetHandle("rlpaddr.nBYTES"),
			ToHash:  comp.Columns.GetHandle("rlpaddr.LC"),
		},
		Info: generic.GenInfoModule{
			HashNum:  comp.Columns.GetHandle("rlpaddr.STAMP"),
			HashLo:   comp.Columns.GetHandle("rlpaddr.DEP_ADDR_LO"),
			HashHi:   comp.Columns.GetHandle("rlpaddr.RAW_ADDR_HI"),
			IsHashLo: comp.Columns.GetHandle("rlpaddr.SELECTOR_KECCAK_RES"),
			IsHashHi: comp.Columns.GetHandle("rlpaddr.SELECTOR_KECCAK_RES"),
		},
	}
}
