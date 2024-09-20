package blobdecompression

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	blob_v0 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0"
	blob_v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"

	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
)

const (
	// It is just a random number. What is important is that it is not used to
	// create another dummy circuit. So don't copy-paste this or at least,
	// change the value if you do.
	MockCircuitID = circuits.MockCircuitIDDecompression
)

// Generates a concrete proof for the decompression of the blob
func Prove(cfg *config.Config, req *Request) (*Response, error) {

	if cfg.BlobDecompression.ProverMode == config.ProverModeDev {
		return dummyProve(cfg, req)
	}

	// Parsing / validating the request
	blobBytes, err := base64.StdEncoding.DecodeString(req.CompressedData)
	if err != nil {
		return nil, fmt.Errorf("could not parse the compressed data: %w", err)
	}

	var (
		xBytes [32]byte
		y      fr381.Element
	)

	if b, err := utils.HexDecodeString(req.ExpectedX); err != nil {
		return nil, fmt.Errorf("could not parse the bytes of the expected x: %w", err)
	} else {
		copy(xBytes[:], b)
	}

	yBytes, err := utils.HexDecodeString(req.ExpectedY)
	if err != nil {
		return nil, fmt.Errorf("could not parse the bytes of the expected y: %w", err)
	}
	y.SetBytes(yBytes)

	// First of all, we need to identify which setup-info to use
	version := blob.GetVersion(blobBytes)
	var (
		circuitID                    circuits.CircuitID
		expectedMaxUsableBytes       int
		expectedMaxUncompressedBytes int
	)
	switch version {
	case 0:
		circuitID = circuits.BlobDecompressionV0CircuitID
		expectedMaxUsableBytes = blob_v0.MaxUsableBytes
		expectedMaxUncompressedBytes = blob_v0.MaxUncompressedBytes
	case 1:
		circuitID = circuits.BlobDecompressionV1CircuitID
		expectedMaxUsableBytes = blob_v1.MaxUsableBytes
		expectedMaxUncompressedBytes = blob_v1.MaxUncompressedBytes
	default:
		return nil, fmt.Errorf("unsupported blob version: %v", version)
	}

	setup, err := circuits.LoadSetup(cfg, circuitID)
	if err != nil {
		return nil, fmt.Errorf("could not load the setup: %w", err)
	}
	dictPath := filepath.Join(cfg.PathForSetup(string(circuitID)), config.DictionaryFileName)

	logrus.Infof("reading the dictionary at %v", dictPath)

	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return nil, fmt.Errorf("error reading the dictionary: %w", err)
	}

	maxUsableBytes, err := setup.Manifest.GetInt("maxUsableBytes")
	if err != nil {
		return nil, fmt.Errorf("missing maxUsableBytes in the setup manifest: %w", err)
	}

	maxUncompressedBytes, err := setup.Manifest.GetInt("maxUncompressedBytes")
	if err != nil {
		return nil, fmt.Errorf("missing maxUncompressedBytes in the setup manifest: %w", err)
	}

	if maxUsableBytes != expectedMaxUsableBytes {
		return nil, fmt.Errorf("invalid maxUsableBytes in the setup manifest: %v, expected %v", maxUsableBytes, expectedMaxUsableBytes)
	}

	if maxUncompressedBytes != expectedMaxUncompressedBytes {
		return nil, fmt.Errorf("invalid maxUncompressedBytes in the setup manifest: %v, expected %v", maxUncompressedBytes, expectedMaxUncompressedBytes)
	}

	// This computes the assignment

	logrus.Infof("computing the circuit's assignment")

	snarkHash, err := utils.HexDecodeString(req.SnarkHash)
	if err != nil {
		return nil, fmt.Errorf("could not parse the snark hash: %w", err)
	}

	assignment, pubInput, _snarkHash, err := blobdecompression.Assign(
		blobBytes,
		dict,
		req.Eip4844Enabled,
		xBytes,
		y,
	)

	if !bytes.Equal(snarkHash, _snarkHash) {
		return nil, fmt.Errorf("blob checksum does not match the one computed by the assigner")
	}

	if err != nil {
		return nil, fmt.Errorf("while generating the assignment: %w", err)
	}

	// Uncomment the following to activate the test.Solver. This is useful to
	// debug when the circuit solving does not pass. It provides more useful
	// information about what was wrong.

	// err := test.IsSolved(
	// 	&assignment,
	// 	&assignment,
	// 	ecc.BLS12_377.ScalarField(),
	// 	test.WithBackendProverOptions(emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField())),
	// )

	// if err != nil {
	// 		panic(err)
	// }

	// This section reads the public parameters. This is a time-consuming part
	// of the process.

	opts := []any{
		emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
		emPlonk.GetNativeVerifierOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
	}

	// This actually runs the compression prover

	logrus.Infof("running the decompression prover")

	proof, err := circuits.ProveCheck(
		&setup,
		assignment,
		opts...,
	)

	if err != nil {
		return nil, fmt.Errorf("while generating the proof: %w", err)
	}

	logrus.Infof("prover successful : generated proof `%++v` for public input `%v`", proof, pubInput.String())

	resp := &Response{
		Request:            *req,
		ProverVersion:      cfg.Version,
		DecompressionProof: circuits.SerializeProofRaw(proof),
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
	}

	resp.Debug.PublicInput = "0x" + pubInput.Text(16)

	return resp, nil

}

// Generates a mocked proof. Assumes the request has already been validated
func dummyProve(cfg *config.Config, req *Request) (*Response, error) {

	// NB: we have not agreed on how to generate the public input with the gnark
	// team yet. So as a default. We return the hash of the request and use the
	// dummy circuit. That way, we get something random-looking. The circuit is
	// differentiated from the dummy circuit that we use to for the light prover.
	// The goal is to ensure that "compression proofs" will be accepted on the
	// light prover contract and vice-versa.

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(req); err != nil {
		return nil, fmt.Errorf("could not encode the request: %w", err)
	}
	input := utils.KeccakHash(buf.Bytes())
	inputF := new(fr.Element).SetBytes(input)

	// Auto-generate the setup (this will not be production settings, obviously)
	// This is only practical because we use the dummy-circuit. The actual
	// circuit would take extremely long time to compile and setup. Beside, we
	// are using a plaintext "toxic-waste".

	srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
	if err != nil {
		return nil, fmt.Errorf("could not create the SRS store: %w", err)
	}

	setup, err := dummy.MakeUnsafeSetup(srsProvider, MockCircuitID, ecc.BLS12_377.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("could not make the setup: %w", err)
	}

	proof := dummy.MakeProof(&setup, *inputF, MockCircuitID)
	resp := &Response{
		Request:            *req,
		ProverVersion:      cfg.Version,
		DecompressionProof: proof,
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
	}

	inputString := utils.HexEncodeToString(input)
	resp.Debug.PublicInput = utils.ApplyModulusBls12377(inputString)

	return resp, nil
}
