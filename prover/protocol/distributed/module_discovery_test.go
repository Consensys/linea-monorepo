package experiment

import "testing"

func TestDiscoveryOnZkEVM(t *testing.T) {

	var (
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 27,
		}
	)

	precompileInitialWizard(zkevm.WizardIOP)

	// The test is to make sure that this function returns
	disc.Analyze(zkevm.WizardIOP)
}
