//go:build !fuzzlight

package glue_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/glue"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/plonk"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark-crypto/ecc"
	cryptoecdsa "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/test"
)

var (
	// dummy values to use instead of extracted data from the execution trace
	dummyPubKey    = [64]byte{0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0xb, 0x7, 0x2, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98, 0x48, 0x3a, 0xda, 0x77, 0x26, 0xa3, 0xc4, 0x65, 0x5d, 0xa4, 0xfb, 0xfc, 0xe, 0x11, 0x8, 0xa8, 0xfd, 0x17, 0xb4, 0x48, 0xa6, 0x85, 0x54, 0x19, 0x9c, 0x47, 0xd0, 0x8f, 0xfb, 0x10, 0xd4, 0xb8}
	dummyTxHash    = [32]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	dummySignature = [65]byte{0xf, 0x20, 0x2f, 0x1b, 0xfd, 0xea, 0x16, 0xb1, 0xed, 0x56, 0x50, 0x7e, 0x2c, 0x3, 0x8c, 0x6e, 0x94, 0xc6, 0xd1, 0x58, 0xfb, 0xae, 0x29, 0x41, 0x3d, 0xd5, 0x78, 0x9a, 0xc9, 0x51, 0x1f, 0x58, 0x89, 0x59, 0xe4, 0x57, 0x1f, 0x7f, 0x1f, 0x4e, 0x15, 0xf4, 0xbb, 0x79, 0x9a, 0x17, 0x65, 0xff, 0x70, 0xe9, 0xdd, 0xa3, 0xca, 0x32, 0x5f, 0x3b, 0x64, 0x45, 0xf7, 0x33, 0x30, 0x86, 0x44, 0xcb, 0x1c}
	// number of transaction signatures verified in tests
	dummyNbTx = 2
	// number of signatures extracted from the execution trace.
	dummyNbECRecover = 4
	dummyNbTotal     = dummyNbTx + dummyNbECRecover
)

const (
	DUMMY_ECDATA_LIMB_NAME = "ECDATA_LIMB"
	SIZE                   = 1 << 12
)

func TestMultiECDSACircuit(t *testing.T) {
	circuit := glue.NewECDSACircuit(dummyNbTotal)
	assignment := glue.NewECDSACircuit(dummyNbTotal)
	txHashes, txPubs, txSigs := dummyTxExtrac()
	assignment.AssignTxWitness(txHashes, txPubs, txSigs)
	pcHashes, pcPubs, pcSigs := dimmyExtractECRecoverCalls(nil)
	assignment.AssignEcDataWitness(pcHashes, pcPubs, pcSigs)

	assert := test.NewAssert(t)
	assert.SolvingSucceeded(
		circuit,
		assignment,
		test.NoFuzzing(),
		test.NoSerialization(),
		test.WithBackends(backend.PLONK),
		test.WithCurves(ecc.BN254),
		test.WithCompileOpts(frontend.WithCapacity(dummyNbTotal*glue.EcRecoverNumConstraints)))
}

func TestWizard(t *testing.T) {
	cmp := wizard.Compile(dummyDefine,
		dummy.Compile,
		// compiler.Arcane(8),
		// vortex.Compile(2),
	)

	proof := wizard.Prove(cmp, dummyProve)

	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}

func dummyDefine(b *wizard.Builder) {
	DUMMY_ECDATA_LIMB := b.RegisterCommit(DUMMY_ECDATA_LIMB_NAME, SIZE)

	glue.RegisterECDSA(b.CompiledIOP, DUMMY_ECDATA_LIMB.Round(), dummyNbTotal, dummyTxExtrac, dimmyExtractECRecoverCalls, glue.WithBatchSize(3), glue.WithPlonkOption(plonk.WithRangecheck(16, 6)))
}

func dummyProve(runtime *wizard.ProverRuntime) {
	runtime.AssignColumn(DUMMY_ECDATA_LIMB_NAME, smartvectors.NewConstant(field.NewElement(3), SIZE))
}

func dimmyExtractECRecoverCalls(comp *wizard.CompiledIOP) (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	// TODO: dummy for now
	for i := 0; i < dummyNbECRecover; i++ {
		txHashes = append(txHashes, dummyTxHash)
		pubKeys = append(pubKeys, dummyPubKey)
		signatures = append(signatures, dummySignature)
	}
	return
}

func dummyTxExtrac() ([][32]byte, [][64]byte, [][65]byte) {
	halfFr := new(big.Int).Sub(emulated.Secp256k1Fr{}.Modulus(), big.NewInt(1))
	halfFr.Div(halfFr, big.NewInt(2))

	var txHashes [][32]byte
	var txPubs [][64]byte
	var txSigs [][65]byte
	for i := 0; i < dummyNbTx; i++ {
		sk, err := cryptoecdsa.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		var msg [32]byte
		_, err = rand.Reader.Read(msg[:])
		if err != nil {
			panic(err)
		}
		var fullSig [65]byte
		for {
			v, r, s, err := sk.SignForRecover(msg[:], nil)
			if err != nil {
				panic(err)
			}
			if halfFr.Cmp(s) <= 0 {
				continue
			}
			r.FillBytes(fullSig[:32])
			s.FillBytes(fullSig[32:64])
			fullSig[64] = byte(v) + 27
			break
		}
		var pub [64]byte
		copy(pub[:], sk.Public().Bytes())
		txHashes = append(txHashes, msg)
		txPubs = append(txPubs, pub)
		txSigs = append(txSigs, fullSig)
	}
	return txHashes, txPubs, txSigs
}
