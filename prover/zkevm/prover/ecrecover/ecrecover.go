package ecrecover

import (
	"fmt"

	cryptoecdsa "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/signature/ecdsa"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// placeholder values to use when we prove less than the upper bound of ECDSA
// signatures. As the circuit is fixed-size, then we need to operate on
// something.
var (
	// public key 1*G
	placeholderPubkey = [64]byte{0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0xb, 0x7, 0x2, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98, 0x48, 0x3a, 0xda, 0x77, 0x26, 0xa3, 0xc4, 0x65, 0x5d, 0xa4, 0xfb, 0xfc, 0xe, 0x11, 0x8, 0xa8, 0xfd, 0x17, 0xb4, 0x48, 0xa6, 0x85, 0x54, 0x19, 0x9c, 0x47, 0xd0, 0x8f, 0xfb, 0x10, 0xd4, 0xb8}
	// 1000000...0000
	placeholderTxHash = [32]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	// valid signature for secret key 1 and tx hash 1 with random nonce
	// valid signature for secret key 1 and tx hash 1 with nonce 2
	placeholderSignature = [65]byte{
		// r part of signature
		0xc6, 0x4, 0x7f, 0x94, 0x41, 0xed, 0x7d, 0x6d, 0x30, 0x45, 0x40, 0x6e, 0x95, 0xc0, 0x7c, 0xd8, 0x5c, 0x77, 0x8e, 0x4b, 0x8c, 0xef, 0x3c, 0xa7, 0xab, 0xac, 0x9, 0xb9, 0x5c, 0x70, 0x9e, 0xe5,
		// s part of signature
		0xe3, 0x82, 0x3f, 0xca, 0x20, 0xf6, 0xbe, 0xb6, 0x98, 0x22, 0xa0, 0x37, 0x4a, 0xe0, 0x3e, 0x6b, 0x8b, 0x93, 0x35, 0x99, 0x1e, 0x1b, 0xee, 0x71, 0xb5, 0xbf, 0x34, 0x23, 0x16, 0x53, 0x70, 0x13,
		// v part of signature (in EVM format, 27 is added to the recovery id)
		0x1b,
	}
)

// Estimation in the number of PLONK constraints needed to verify a signature
// this estimate takes a margin of around ~10% compared to what we have measured
const EcRecoverNumConstraints int = 1_800_000

// TxSignatureExtractor extracts the transaction hashes, public keys and
// signatures at proving time. The lengths of the outputs must match. The
// extracted public keys must be raw, i.e. not hashed into Ethereum address.
type TxSignatureExtractor func() (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte)

// Function that returns no transaction signature that can be used to generate
// the wizard without intention of actually proving it directly.
func NoTxSignatures() (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	return [][32]byte{}, [][64]byte{}, [][65]byte{}
}

// TraceSignatureExtractor extracts the prehashes, public keys and signatures
// from the execution trace at proving time. The lengths of the outputs must
// match. The extracted public keys must be raw, i.e. not hashed into Ethereum
// address. The default implementation from the EC_DATA module from the ZKEVM
// execution trace is is [DefaultTraceExtractor].
type TraceSignatureExtractor func(*wizard.CompiledIOP) (prehashed [][32]byte, pubKeys [][64]byte, signatures [][65]byte)

// RecoverableSignature is an ECDSA signature with recover information V. V must
// be 27 or 28 (like Ethereum).
type RecoverableSignature struct {
	ecdsa.Signature[emulated.Secp256k1Fr]
	V frontend.Variable
}

// fromRawSignature returns [RecoverableSignature] from the bytes. We assumme
// byte format, where first 32 bytes are big-endian R, next 32 bytes big-endian
// S and last byte V value.
func fromRawSignature(signature [65]byte) RecoverableSignature {
	return RecoverableSignature{
		Signature: ecdsa.Signature[emulated.Secp256k1Fr]{
			R: emulated.ValueOf[emulated.Secp256k1Fr](signature[0:32]),
			S: emulated.ValueOf[emulated.Secp256k1Fr](signature[32:64]),
		},
		V: signature[64],
	}
}

// MultiECDSACircuit verifies multiple ECDSA signatures. The number of total
// verifiable signatures is bounded by the sizes of the fields in the structure.
// To assign witness values, use the methods [MultiECDSACircuit.AssignTxWitness]
// and [MultiECDSACircuit.AssignEcDataWitness]. The type implements
// [frontend.Circuit].
type MultiECDSACircuit struct {
	// StrictCheck is a boolean slice indicating if the S value in the
	// signature:
	//  - must be less than half the scalar modulus if 1
	//  - must be less than the scalar modulus if 0
	StrictCheck []frontend.Variable
	// Msgs is a slice of the prehashed messages to be verified
	Msgs []emulated.Element[emulated.Secp256k1Fr]
	// Signatures is a slice of signatures to verify ECDSA signatures with.
	Signatures []RecoverableSignature
	// Pubkeys is a slice of raw ECDSA public key points to verify for.
	Pubkeys []ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]
	// FailureCheck is a boolean slice indicating if the inputs are invalid
	// (recovered public key is 0 or have quadratic non-residue during
	// commitment computation)
	FailureCheck []frontend.Variable

	witnessLength int
}

// NewECDSACircuit initializes [MultiECDSACircuit] allowing verifying up to numECDSAs signatures.
func NewECDSACircuit(numECDSAs int) *MultiECDSACircuit {
	circuit := &MultiECDSACircuit{
		StrictCheck:  make([]frontend.Variable, numECDSAs),
		Msgs:         make([]emulated.Element[emulated.Secp256k1Fr], numECDSAs),
		Signatures:   make([]RecoverableSignature, numECDSAs),
		Pubkeys:      make([]ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr], numECDSAs),
		FailureCheck: make([]frontend.Variable, numECDSAs),
	}
	// we initialize all witness to place holder values. They will be
	// overwritten by actual hashes/pubkeys/signatures by calls to
	// [MultiECDSACircuit.AssignTxWitness] and
	// [MultiECDSACircuit.AssignEcDataWitness]
	for i := 0; i < numECDSAs; i++ {
		circuit.assignWitness([][32]byte{placeholderTxHash}, [][64]byte{placeholderPubkey}, [][65]byte{placeholderSignature}, 0, 0)
	}
	// reset the witnessLength to zero for correct tracking afterwards.
	circuit.witnessLength = 0
	return circuit
}

// Define implements multi-ECDSA verification logic. Required for implementing [frontend.Circuit].
func (c *MultiECDSACircuit) Define(api frontend.API) error {

	logrus.Debugf("ECDSA in PLONK : Starting the define")

	if len(c.StrictCheck) != len(c.Msgs) ||
		len(c.StrictCheck) != len(c.Signatures) ||
		len(c.StrictCheck) != len(c.Pubkeys) {
		return fmt.Errorf("mismatching lengths")
	}
	ec, err := sw_emulated.New[emulated.Secp256k1Fp, emulated.Secp256k1Fr](api, sw_emulated.GetSecp256k1Params())
	if err != nil {
		return fmt.Errorf("new SW: %w", err)
	}
	for i := range c.StrictCheck {
		logrus.Debugf("ECDSA in PLONK : Define - check %v/%v", i, len(c.StrictCheck))
		recoveredPub := evmprecompiles.ECRecover(
			api,
			c.Msgs[i],
			c.Signatures[i].V, c.Signatures[i].R, c.Signatures[i].S,
			c.StrictCheck[i],
			c.FailureCheck[i],
		)
		ec.AssertIsEqual(recoveredPub, (*sw_emulated.AffinePoint[emulated.Secp256k1Fp])(&c.Pubkeys[i]))
	}

	logrus.Debugf("ECDSA in PLONK : Define - done")

	return nil
}

// assignWitness assignes the witness. The method extracts the corresponding
// bytes from the inputs and initializes the instances the underlying circuit
// expects. The argumnet strict defines if we perform for the inputs strict
// check on s or not (less than half the modulus).
func (c *MultiECDSACircuit) assignWitness(txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte, strict int, failureCheck int) {

	if len(txHashes) != len(pubKeys) || len(txHashes) != len(signatures) {
		panic("mismatching input lengths")
	}
	if len(txHashes)+c.witnessLength > len(c.Msgs) {
		panic("too many witness elements assigned")
	}
	ptr := c.witnessLength
	for i := range txHashes {
		c.Msgs[ptr+i] = emulated.ValueOf[emulated.Secp256k1Fr](cryptoecdsa.HashToInt(txHashes[i][:]))
		c.Pubkeys[ptr+i] = ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](pubKeys[i][0:32]),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](pubKeys[i][32:64]),
		}
		c.Signatures[ptr+i] = fromRawSignature(signatures[i])
		c.StrictCheck[ptr+i] = strict
		c.FailureCheck[ptr+i] = failureCheck
	}
	c.witnessLength += len(txHashes)
}

// AssignTxWitness assigns transaction verification data to the witness by
// setting strict check flag.
func (c *MultiECDSACircuit) AssignTxWitness(txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	c.assignWitness(txHashes, pubKeys, signatures, 1, 0)
}

// AssignEcDataWitness assigns ECRecover precompile verification data to the
// witness by leaving strict check flag unset.
func (c *MultiECDSACircuit) AssignEcDataWitness(txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	c.assignWitness(txHashes, pubKeys, signatures, 0, 0)
}

// AssignInvalidTxWitness assigns invalid transaction verification data to the
// witness by setting strict check flag and failure check flag.
func (c *MultiECDSACircuit) AssignInvalidTxWitness(txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	c.assignWitness(txHashes, pubKeys, signatures, 0, 1)
}

// RegisterECDSA registers multi-ECDSA signature verification circuit in the
// Wizard. It first creates the corresponding gnark circuits, extracts the data
// from transactions using txSigExtractor, then the data from the execution
// traces using traceSigExtractor, assigns the public witness and then calls
// PLONK verifier in Wizard.
func RegisterECDSA(comp *wizard.CompiledIOP, round int, numECDSA int, txSigExtractor TxSignatureExtractor, traceSigExtractor, traceSigInvalidExtractor TraceSignatureExtractor, ops ...Option) {

	logrus.Debugf("ECDSA - registers ECDSA - starting")

	// Applies the options
	options := &option{}
	for _, op := range ops {
		op(options)
	}

	NbEcdsaPerCircuit := numECDSA // default value

	//  takes the value from the options if available
	if options.BatchSize > 0 {
		NbEcdsaPerCircuit = options.BatchSize
	}

	// ensures the batches are no larger than the total number of ECDSA circuits
	if NbEcdsaPerCircuit > numECDSA {
		NbEcdsaPerCircuit = numECDSA
	}

	numBatches := utils.DivCeil(numECDSA, NbEcdsaPerCircuit)

	txHashes, txPubKeys, txSigs := [][32]byte{}, [][64]byte{}, [][65]byte{}
	pcHashes, pcPubs, pcSigs := [][32]byte{}, [][64]byte{}, [][65]byte{}
	pcInvalidHashes, pcInvalidPubs, pcInvalidSigs := [][32]byte{}, [][64]byte{}, [][65]byte{}

	circuit := NewECDSACircuit(NbEcdsaPerCircuit)

	// gathers all the signatures from the extractor
	if txSigExtractor != nil {
		txHashes, txPubKeys, txSigs = txSigExtractor()
	}

	// gathers all the signatures from the traces extractor this is
	// currently plug in function. After ECDATA module has finalized
	// then will perform actual extraction from the execution trace.
	if traceSigExtractor != nil {
		pcHashes, pcPubs, pcSigs = traceSigExtractor(comp)
	}

	// gathers all the invalid signatures from the traces extractor
	if traceSigInvalidExtractor != nil {
		pcInvalidHashes, pcInvalidPubs, pcInvalidSigs = traceSigInvalidExtractor(comp)
	}
	if len(txHashes)+len(pcHashes)+len(pcInvalidHashes) > numECDSA {
		panic(fmt.Sprintf("requested number of signatures %d more than verification capacity %d", len(txHashes)+len(pcHashes)+len(pcInvalidHashes), numECDSA))
	}

	// Then use the place-holder signatures to reach the number of ECDSA that we want
	for i := len(pcHashes) + len(txHashes) + len(pcInvalidHashes); i < numECDSA; i++ {
		txHashes = append(txHashes, placeholderTxHash)
		txPubKeys = append(txPubKeys, placeholderPubkey)
		txSigs = append(txSigs, placeholderSignature)
	}

	// prepare the different assigners
	assignFns := make([]func() frontend.Circuit, numBatches)
	for i := range assignFns {
		assignFns[i] = func() frontend.Circuit {
			logrus.Debugf("ECDSA - registers ECDSA - Witness assignment for circuit %d", i)
			assignment := NewECDSACircuit(NbEcdsaPerCircuit)
			nbAssigned := 0

			// assign the tx hashes in priority
			if i*NbEcdsaPerCircuit < len(txHashes) {
				nbAssignedFromTxs := utils.Min(NbEcdsaPerCircuit, len(txHashes[i*NbEcdsaPerCircuit:]))
				assignment.AssignTxWitness(
					txHashes[i*NbEcdsaPerCircuit:i*NbEcdsaPerCircuit+nbAssignedFromTxs],
					txPubKeys[i*NbEcdsaPerCircuit:i*NbEcdsaPerCircuit+nbAssignedFromTxs],
					txSigs[i*NbEcdsaPerCircuit:i*NbEcdsaPerCircuit+nbAssignedFromTxs],
				)
				nbAssigned += nbAssignedFromTxs
			}

			// and then assign the ec recover signatures
			if nbAssigned < NbEcdsaPerCircuit {
				startFrom := i*NbEcdsaPerCircuit + nbAssigned - len(txHashes)
				nbAssignedFromEc := utils.Min(NbEcdsaPerCircuit-nbAssigned, len(pcHashes[startFrom:]))
				assignment.AssignEcDataWitness(
					pcHashes[startFrom:startFrom+nbAssignedFromEc],
					pcPubs[startFrom:startFrom+nbAssignedFromEc],
					pcSigs[startFrom:startFrom+nbAssignedFromEc],
				)
				nbAssigned += len(pcHashes[startFrom : startFrom+nbAssignedFromEc])
			}

			// and then assign the invalid ec recover signatures
			if nbAssigned < NbEcdsaPerCircuit {
				startFrom := i*NbEcdsaPerCircuit + nbAssigned - len(txHashes) - len(pcHashes)
				nbAssignedInvalidFromEc := utils.Min(NbEcdsaPerCircuit-nbAssigned, len(pcInvalidHashes[startFrom:]))
				assignment.AssignInvalidTxWitness(
					pcInvalidHashes[startFrom:startFrom+nbAssignedInvalidFromEc],
					pcInvalidPubs[startFrom:startFrom+nbAssignedInvalidFromEc],
					pcInvalidSigs[startFrom:startFrom+nbAssignedInvalidFromEc],
				)
			}

			if len(assignment.Msgs) != NbEcdsaPerCircuit {
				utils.Panic(
					"Should have assigned %d msgs, but assigned %v",
					NbEcdsaPerCircuit,
					len(assignment.Msgs),
				)
			}

			logrus.Debugf("ECDSA - registers ECDSA - Witness assignment for circuit %d - done", i)
			return assignment
		}

	}

	logrus.Debugf("ECDSA - registers ECDSA - running the PLONK check")
	plonk.PlonkCheck(comp, "ECDSA", round, circuit, assignFns, plonk.WithRangecheck(21, 4, true))
	logrus.Debugf("ECDSA - registers ECDSA - starting")
}

// DefaultTraceExtractor extracts calls to ECRecover precompile from the ZKEVM
// execution trace. Currently not implemented and panics if called.
func DefaultTraceExtractor(comp *wizard.CompiledIOP) (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	panic("EC_DATA signature extractor not implemented")
}
