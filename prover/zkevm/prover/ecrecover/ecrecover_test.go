//go:build !fuzzlight

package ecrecover_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	cryptoecdsa "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/test"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/ecrecover"
	"github.com/sirupsen/logrus"
)

var (
	// dummy values to use instead of extracted data from the execution trace
	// 1*G
	dummyPubKey = [64]byte{0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0xb, 0x7, 0x2, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98, 0x48, 0x3a, 0xda, 0x77, 0x26, 0xa3, 0xc4, 0x65, 0x5d, 0xa4, 0xfb, 0xfc, 0xe, 0x11, 0x8, 0xa8, 0xfd, 0x17, 0xb4, 0x48, 0xa6, 0x85, 0x54, 0x19, 0x9c, 0x47, 0xd0, 0x8f, 0xfb, 0x10, 0xd4, 0xb8}
	// 1000000...0000
	dummyTxHash = [32]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	// valid signature for secret key 1 and tx hash 1 with nonce 2
	dummySignature = [65]byte{
		// r part of signature
		0xc6, 0x4, 0x7f, 0x94, 0x41, 0xed, 0x7d, 0x6d, 0x30, 0x45, 0x40, 0x6e, 0x95, 0xc0, 0x7c, 0xd8, 0x5c, 0x77, 0x8e, 0x4b, 0x8c, 0xef, 0x3c, 0xa7, 0xab, 0xac, 0x9, 0xb9, 0x5c, 0x70, 0x9e, 0xe5,
		// s part of signature
		0xe3, 0x82, 0x3f, 0xca, 0x20, 0xf6, 0xbe, 0xb6, 0x98, 0x22, 0xa0, 0x37, 0x4a, 0xe0, 0x3e, 0x6b, 0x8b, 0x93, 0x35, 0x99, 0x1e, 0x1b, 0xee, 0x71, 0xb5, 0xbf, 0x34, 0x23, 0x16, 0x53, 0x70, 0x13,
		// v part of signature (in EVM format, 27 is added to the recovery id)
		0x1b,
	}
	invalidSignature = [65]byte{
		// r part of signature
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xfc, 0x2e,
		// s part of signature
		0x44, 0xc8, 0x19, 0xd6, 0xb9, 0x71, 0xe4, 0x56, 0x56, 0x2f, 0xef, 0xc2, 0x40, 0x85, 0x36, 0xbd, 0xfd, 0x95, 0x67, 0xee, 0x1c, 0x6c, 0x7d, 0xd2, 0xa7, 0x7, 0x66, 0x25, 0x95, 0x3a, 0x18, 0x59,
		// v part of signature (in EVM format, 27 is added to the recovery id)
		0x1c,
	}
	// number of transaction signatures verified in tests
	dummyNbTx = 2
	// number of signatures extracted from the execution trace.
	dummyNbECRecover = 4
	// number of invalid signatures verified in tests
	dummyNbInvalidInf = 1
	dummyNbInvalidQNR = 1

	dummyNbTotal = dummyNbTx + dummyNbECRecover + dummyNbInvalidInf + dummyNbInvalidQNR
)

const (
	DUMMY_ECDATA_LIMB_NAME = "ECDATA_LIMB"
	SIZE                   = 1 << 12
)

func TestMultiECDSACircuit(t *testing.T) {

	// Skipped because unused
	// t.Skip()

	circuit := ecrecover.NewECDSACircuit(dummyNbTotal)
	assignment := ecrecover.NewECDSACircuit(dummyNbTotal)
	txHashes, txPubs, txSigs := dummyTxExtrac()
	assignment.AssignTxWitness(txHashes, txPubs, txSigs)
	pcHashes, pcPubs, pcSigs := dummyExtractECRecoverCalls(nil)
	assignment.AssignEcDataWitness(pcHashes, pcPubs, pcSigs)

	assert := test.NewAssert(t)
	assert.SolvingSucceeded(
		circuit,
		assignment,
		test.NoFuzzing(),
		test.NoSerializationChecks(),
		test.WithBackends(backend.PLONK),
		test.WithCurves(ecc.BLS12_377),
		test.WithCompileOpts(frontend.WithCapacity(dummyNbTotal*ecrecover.EcRecoverNumConstraints)))
}

func TestWizard(t *testing.T) {

	// t.SkipNow()

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

	ecrecover.RegisterECDSA(
		b.CompiledIOP,
		DUMMY_ECDATA_LIMB.Round(),
		dummyNbTotal,
		dummyTxExtrac,
		dummyExtractECRecoverCalls,
		dummyExtractInvalidECRecoverCalls,
		ecrecover.WithBatchSize(3),
		ecrecover.WithPlonkOption(plonk.WithRangecheck(16, 6, true)),
	)
}

func dummyProve(runtime *wizard.ProverRuntime) {
	runtime.AssignColumn(DUMMY_ECDATA_LIMB_NAME, smartvectors.NewConstant(field.NewElement(3), SIZE))
}

func dummyExtractECRecoverCalls(comp *wizard.CompiledIOP) (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	// TODO: dummy for now
	for i := 0; i < dummyNbECRecover; i++ {
		txHashes = append(txHashes, dummyTxHash)
		pubKeys = append(pubKeys, dummyPubKey)
		signatures = append(signatures, dummySignature)
	}
	return
}

func dummyExtractInvalidECRecoverCalls(comp *wizard.CompiledIOP) (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
	zeroPubKey := [64]byte{}
	for i := 0; i < dummyNbInvalidQNR; i++ {
		txHashes = append(txHashes, dummyTxHash)
		pubKeys = append(pubKeys, zeroPubKey)
		signatures = append(signatures, invalidSignature)
	}
	var sk cryptoecdsa.PrivateKey
	for i := 0; i < dummyNbInvalidInf; i++ {
		v, r, s, err := sk.SignForRecover(dummyTxHash[:], nil)
		if err != nil {
			panic(err)
		}
		var signature [65]byte
		r.FillBytes(signature[:32])
		s.FillBytes(signature[32:64])
		signature[64] = byte(v) + 27
		txHashes = append(txHashes, dummyTxHash)
		pubKeys = append(pubKeys, zeroPubKey)
		signatures = append(signatures, signature)
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

func BenchmarkECRecoverWizard(b *testing.B) {

	logrus.SetLevel(logrus.FatalLevel)
	numberSigs := []int{48, 96, 192, 388, 482}

	for _, nbSig := range numberSigs {

		benchName := fmt.Sprintf("%v-signatures", nbSig)
		b.Run(benchName, func(b *testing.B) {

			for _b := 0; _b < b.N; _b++ {

				// @alex: we don't aim at assigning this wizard. We just want to know
				// the number of columns.
				def := func(b *wizard.Builder) {
					ecrecover.RegisterECDSA(
						b.CompiledIOP,
						0,
						nbSig,
						func() (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) { return nil, nil, nil },
						func(ci *wizard.CompiledIOP) (prehashed [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
							return nil, nil, nil
						},
						func(ci *wizard.CompiledIOP) (prehashed [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {
							return nil, nil, nil
						},
						ecrecover.WithBatchSize(3),
						ecrecover.WithPlonkOption(plonk.WithRangecheck(16, 6, true)),
					)
				}

				// @alex: the compilation purposefully does not use compilation
				// beyond the lookup expansion.
				comp := wizard.Compile(
					def,
					specialqueries.RangeProof,
					specialqueries.CompileFixedPermutations,
					permutation.CompileGrandProduct,
					lookup.CompileLogDerivative,
					innerproduct.Compile,
				)

				if _b > 0 {
					continue
				}

				var (
					totalWeight = 0
					allColIDs   = comp.Columns.AllKeys()
				)

				for _, colID := range allColIDs {
					totalWeight += comp.Columns.GetSize(colID)
				}

				b.ReportMetric(float64(totalWeight), "cells")
				b.ReportMetric(float64(nbSig), "signatures")
			}
		})
	}
}
