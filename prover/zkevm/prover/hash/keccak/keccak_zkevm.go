// The keccak package accumulates the providers from different zkEVM modules,
//
//	and proves the hash consistency over the unified provider.
//
// The provider encodes the inputs and outputs of the hash from different modules.
package keccak

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	gen_acc "github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak/acc_module"
)

type KeccakZkEVMInput struct {
	Settings *Settings
	// the list of providers, encoding inputs and outputs of hash
	Providers []generic.GenericByteModule
}
type KeccakZkEVM struct {
	Settings *Settings

	// The [wizard.ProverAction] for submodules.
	pa_accData wizard.ProverAction
	pa_accInfo wizard.ProverAction
	pa_keccak  wizard.ProverAction
}

func NewKeccakZkEVM(comp *wizard.CompiledIOP, inp KeccakZkEVMInput) *KeccakZkEVM {

	// create the list of  [generic.GenDataModule] and [generic.GenInfoModule]
	var gdm []generic.GenDataModule
	var gim []generic.GenInfoModule

	for i := range inp.Providers {
		gdm = append(gdm, inp.Providers[i].Data)
		gim = append(gim, inp.Providers[i].Info)
	}

	var (
		inpAcc = gen_acc.GenericAccumulatorInputs{
			MaxNumKeccakF: inp.Settings.MaxNumKeccakf,
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
			MaxNumKeccakF: inp.Settings.MaxNumKeccakf,
		}
		keccak = NewKeccakSingleProvider(comp, keccakInp)
	)

	res := &KeccakZkEVM{
		pa_accData: accData,
		pa_accInfo: accInfo,
		pa_keccak:  keccak}
	return res

}

func (k *KeccakZkEVM) Run(run *wizard.ProverRuntime) {
	k.pa_accData.Run(run)
	k.pa_accInfo.Run(run)
	k.pa_keccak.Run(run)
}
