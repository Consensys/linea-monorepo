package invalidity

import (
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	invalidity "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/invalidity_pi"
)

const MAX_L2_LOGS = 16

// BadPrecompileCircuit defines the circuit for the transaction with a bad precompile.
type BadPrecompileCircuit struct {
	// The wizard verifier circuit - this is the main witness
	WizardVerifier wizard.VerifierCircuit `gnark:",secret"`

	// These fields are derived from wizard public inputs
	// They are excluded from witness generation via gnark:"-" tag
	stateRootHash    [2]frontend.Variable `gnark:"-"`
	txHash           [2]frontend.Variable `gnark:"-"`
	fromAddress      frontend.Variable    `gnark:"-"`
	hasBadPrecompile frontend.Variable    `gnark:"-"`
	NbL2Logs         frontend.Variable    `gnark:"-"`

	// Recipient address (not constrained in this subcircuit, but included in public input)
	ToAddress      frontend.Variable
	InvalidityType frontend.Variable
}

func (circuit *BadPrecompileCircuit) Allocate(config Config) {

	wverifier := wizard.AllocateWizardCircuit(config.Zkevm.InitialCompiledIOP, 0, true)
	circuit.WizardVerifier = *wverifier
}

func (circuit *BadPrecompileCircuit) Define(api frontend.API) error {

	circuit.WizardVerifier.BLSFS = fiatshamir.NewGnarkFSBLS12377(api)
	circuit.WizardVerifier.Verify(api)

	// Extract public inputs of the wizard circuit first
	pie := getInvalidityPIExtractor(&circuit.WizardVerifier)
	circuit.checkPublicInputs(api, &circuit.WizardVerifier, pie)

	// check that invalidity type is valid, it should be 2 or 3
	// binaryType = 0 when BadPrecompile (type 2), binaryType = 1 when TooManyLogs (type 3)
	binaryType := api.Sub(circuit.InvalidityType, 2)
	api.AssertIsBoolean(binaryType)

	// check that hasBadPrecompile is non-zero, if invalidityType == 2 (BadPrecompile)
	// When binaryType=0: (1-0)*hasBadPrecompile + 0 = hasBadPrecompile != 0 ✓
	// When binaryType=1: (1-1)*hasBadPrecompile + 1 = 1 != 0 ✓ (always passes)
	api.AssertIsDifferent(
		api.Add(
			api.Mul(api.Sub(1, binaryType), circuit.hasBadPrecompile),
			binaryType),
		0)

	// check that NbL2Logs is greater than MAX_L2_LOGS, if invalidityType == 3 (TooManyLogs)
	// When binaryType=1: 17 <= 1*NbL2Logs + 0 = NbL2Logs ✓
	// When binaryType=0: 17 <= 0*NbL2Logs + 17 = 17 ✓ (always passes)
	api.AssertIsLessOrEqual(MAX_L2_LOGS+1,
		api.Add(
			api.Mul(binaryType, circuit.NbL2Logs),
			api.Mul(api.Sub(1, binaryType), MAX_L2_LOGS+1),
		),
	)

	return nil
}

// checkPublicInputs checks the public inputs of the wizard circuit
func (circuit *BadPrecompileCircuit) checkPublicInputs(api frontend.API, wvc *wizard.VerifierCircuit, pie *invalidity.InvalidityPIExtractor) {
	// Extract StateRootHash (8 KoalaBear elements) and reconstruct
	extrStateRootHashWords := getPublicInputArr(api, wvc, pie.StateRootHash[:])
	circuit.stateRootHash[0] = combine32BitLimbs(api, extrStateRootHashWords[:4])
	circuit.stateRootHash[1] = combine32BitLimbs(api, extrStateRootHashWords[4:])

	// Extract TxHash (16 BE limbs) and reconstruct using compress.ReadNum
	extrTxHashLimbs := getPublicInputArr(api, wvc, pie.TxHash[:])

	circuit.txHash[0] = combine16BitLimbs(api, extrTxHashLimbs[:8])
	circuit.txHash[1] = combine16BitLimbs(api, extrTxHashLimbs[8:])
	// Extract FromAddress (10 BE limbs) and reconstruct
	extrFromAddressLimbs := getPublicInputArr(api, wvc, pie.FromAddress[:])
	circuit.fromAddress = combine16BitLimbs(api, extrFromAddressLimbs)

	// Extract scalar public inputs
	circuit.hasBadPrecompile = getPublicInput(api, wvc, pie.HasBadPrecompile)
	circuit.NbL2Logs = getPublicInput(api, wvc, pie.NbL2Logs)

}

// getInvalidityPIExtractor retrieves the InvalidityPIExtractor from wizard circuit metadata
// This follows the same pattern as execution circuit (pi_wizard_extraction.go:287-296)
func getInvalidityPIExtractor(wvc *wizard.VerifierCircuit) *invalidity.InvalidityPIExtractor {
	extraData, extraDataFound := wvc.Spec.ExtraData[invalidity.InvalidityPIExtractorMetadata]
	if !extraDataFound {
		panic("invalidity PI extractor not found")
	}
	pie, ok := extraData.(*invalidity.InvalidityPIExtractor)
	if !ok {
		panic("invalidity PI extractor not of the right type: " + reflect.TypeOf(extraData).String())
	}
	return pie
}

// getPublicInputArr returns a slice of values from the public input
// Reused from execution circuit (pi_wizard_extraction.go:313-320)
func getPublicInputArr(api frontend.API, wvc *wizard.VerifierCircuit, pis []wizard.PublicInput) []frontend.Variable {
	res := make([]frontend.Variable, len(pis))
	for i := range pis {
		r := wvc.GetPublicInput(api, pis[i].Name)
		res[i] = r.Native()
	}
	return res
}

// getPublicInput returns a value from the public input
// Reused from execution circuit (pi_wizard_extraction.go:323-326)
func getPublicInput(api frontend.API, wvc *wizard.VerifierCircuit, pi wizard.PublicInput) frontend.Variable {
	r := wvc.GetPublicInput(api, pi.Name)
	return r.Native()
}

// Assign assigns the inputs to the circuit
func (circuit *BadPrecompileCircuit) Assign(assi AssigningInputs) {
	circuit.WizardVerifier = *wizard.AssignVerifierCircuit(assi.Zkevm.InitialCompiledIOP, assi.ZkevmWizardProof, 0, true)
	circuit.InvalidityType = int(assi.InvalidityType) // cast to int for gnark witness

	// Assign ToAddress (not constrained, but flows into public input hash)
	circuit.ToAddress = assi.Transaction.To()[:]

}

// FunctionalPIQGnark returns the subcircuit-derived functional public inputs
func (c *BadPrecompileCircuit) FunctionalPIQGnark() FunctinalPIQGnark {
	return FunctinalPIQGnark{
		FromAddress:   c.fromAddress,
		TxHash:        c.txHash,
		StateRootHash: c.stateRootHash,
		ToAddress:     c.ToAddress,
	}
}
