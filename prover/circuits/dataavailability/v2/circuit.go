package v2

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	gcposeidon2permutation "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/std/compress/lzss"
	gkrposeidon2 "github.com/consensys/gnark/std/hash/poseidon2/gkr-poseidon2"
	poseidon2permutation "github.com/consensys/gnark/std/permutation/poseidon2/gkr-poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/dataavailability/config"
	"github.com/consensys/linea-monorepo/prover/circuits/dataavailability/publicinput"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"

	"github.com/consensys/gnark-crypto/ecc"
	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	gcHash "github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/compress"
	snarkHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"

	blob "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v2"
)

// Circuit used to prove data compression and blob submission related operations
// It performs the following operations:
//   - Proves the compression
//   - Prove the equivalence of the compressed data with the one claimed on
//     L1. Through an evaluation check
//   - Prove the equivalence of the decompressed datastream with the stream
//     used to prove EVM execution through hashing using a SNARK friendly
//     hash function.
//
// The circuit has a single public input derived by mimc hashing the compressed
// data (or rather its mimc hash), the hash of the uncompressed data stream and
// the X and Y values that are computed on L1 and that serves proving the
// equivalence of the compressed data stream in the witness and the one claimed
// on L1.
type Circuit struct {
	config.CircuitSizes

	// The dictionary used in the compression algorithm
	Dict []frontend.Variable

	// The uncompressed and compressed data corresponds to the data that is
	// made available on L1 when we submit a blob of transactions. The circuit
	// proves that these two correspond to the same data. The uncompressed
	// data is then passed to the EVM execution circuit.
	BlobBytes []frontend.Variable

	// The final public input. It is the hash of
	// 	- the hash of the uncompressed message.
	// 	- the hash of the compressed message. aka the SnarkHash on the contract
	// 	- the X and Y evaluation points claimed on the contract
	PublicInput frontend.Variable `gnark:",public"`

	FuncPI FunctionalPublicInputSnark
}

// FunctionalPublicInputQSnark the "unique" portion of the functional public input that cannot be inferred from other circuits in the same aggregation batch
type FunctionalPublicInputQSnark struct {
	// The X and Y evaluation point that are "claimed" by the contract. The
	// circuit will verify that the polynomial describing the compressed data
	// does evaluate to Y when evaluated at point X. This needed to ensure that
	// both the keccak hash of the compressed data (claimed on the contract) and
	// the blob-hash are correct.
	X               [32]frontend.Variable // unreduced value
	Y               [2]frontend.Variable  // Y[1] holds the less significant 16 bytes
	SnarkHash       frontend.Variable
	Eip4844Enabled  frontend.Variable
	NbBatches       frontend.Variable
	BatchSumsNative []frontend.Variable // batch checksums, proven over the current field
}

type FunctionalPublicInputSnark struct {
	FunctionalPublicInputQSnark
	BatchSumsSmallField []frontend.Variable // batch checksums over small field; unverified
	BatchSumsTotal      []frontend.Variable // H(batchSumNative, batchSumSmallField)
}

type BatchSums struct {
	Native     []byte // batch checksums over the current field
	SmallField []byte // batch checksums over small field; unverified by the blob circuit
	Total      []byte // the batch evaluated at H(batchSumNative, batchSumSmallField)
}

func (bs BatchSums) EvaluationPoint() ([]byte, error) {
	return gcposeidon2permutation.
		NewDefaultPermutation().
		Compress(bs.Native, bs.SmallField)
}

type FunctionalPublicInput struct {
	X              [32]byte
	Y              [2][16]byte
	SnarkHash      []byte
	Eip4844Enabled bool
	BatchSums      []BatchSums // BatchSums checksums for batches of execution data - each corresponding to one exec proof
}

// Check checks that values are within range
func (i *FunctionalPublicInputQSnark) Check(api frontend.API, config config.CircuitSizes) error {
	if len(i.BatchSumsNative) != config.MaxNbBatches {
		return fmt.Errorf("native batch sums capacity: expected %d, got %d", config.MaxNbBatches, len(i.BatchSumsNative))
	}
	rc := rangecheck.New(api)
	for j := range i.X {
		rc.Check(i.X[j], 8)
	}
	rc.Check(i.Y[0], 128)
	rc.Check(i.Y[1], 128)
	api.AssertIsLessOrEqual(i.NbBatches, config.MaxNbBatches) // if too big it can turn "negative" and compromise the interconnection logic
	return nil
}

func (i *FunctionalPublicInputSnark) Check(api frontend.API, config config.CircuitSizes) error {
	if len(i.BatchSumsTotal) != config.MaxNbBatches {
		return fmt.Errorf("batch sums capacity: expected %d, got %d", config.MaxNbBatches, len(i.BatchSumsTotal))
	}
	if len(i.BatchSumsSmallField) != config.MaxNbBatches {
		return fmt.Errorf("small-field batch sums capacity: expected %d, got %d", config.MaxNbBatches, len(i.BatchSumsSmallField))
	}
	return i.FunctionalPublicInputQSnark.Check(api, config)
}

func (i *FunctionalPublicInput) ToSnarkType(maxNbBatches int) (FunctionalPublicInputSnark, error) {
	res := FunctionalPublicInputSnark{
		FunctionalPublicInputQSnark: FunctionalPublicInputQSnark{
			Y:              [2]frontend.Variable{i.Y[0], i.Y[1]},
			SnarkHash:      i.SnarkHash,
			Eip4844Enabled: utils.Ite(i.Eip4844Enabled, 1, 0),
			NbBatches:      len(i.BatchSums),
		},
	}
	utils.Copy(res.X[:], i.X[:])

	if len(i.BatchSums) > maxNbBatches {
		return res, fmt.Errorf("too many batches: expected at most %d, got %d", maxNbBatches, len(i.BatchSums))
	}
	res.BatchSumsNative = make([]frontend.Variable, maxNbBatches)
	res.BatchSumsSmallField = make([]frontend.Variable, maxNbBatches)
	res.BatchSumsTotal = make([]frontend.Variable, maxNbBatches)

	for j := range i.BatchSums {
		res.BatchSumsNative[j] = i.BatchSums[j].Native
		res.BatchSumsSmallField[j] = i.BatchSums[j].SmallField
		res.BatchSumsTotal[j] = i.BatchSums[j].Total
	}

	for j := len(i.BatchSums); j < maxNbBatches; j++ {
		res.BatchSumsNative[j] = 0
		res.BatchSumsSmallField[j] = 0
		res.BatchSumsTotal[j] = 0
	}

	return res, nil
}

func (i *FunctionalPublicInput) Sum() ([]byte, error) {
	hsh := gcHash.POSEIDON2_BLS12_377.New()

	for n := range i.BatchSums {
		// The evaluation point is a hash of the native and small field batch sums.
		if evalPt, err := i.BatchSums[n].EvaluationPoint(); err != nil {
			return nil, err
		} else {
			hsh.Write(evalPt)
		}
		hsh.Write(i.BatchSums[n].Total)
	}

	batchesSum := hsh.Sum(nil)
	hsh.Reset()

	hsh.Write(i.X[:16])
	hsh.Write(i.X[16:])
	hsh.Write(i.Y[0][:])

	// To elimintate a hash permutation, incorporate Eip4844Enabled flag into Y[1], which
	// is never full
	var buf [len(i.Y[1]) + 1]byte
	copy(buf[:], i.Y[1][:])
	buf[len(i.Y[1])] = utils.Ite(i.Eip4844Enabled, byte(1), 0)
	hsh.Write(buf[:])

	hsh.Write(i.SnarkHash)
	hsh.Write(batchesSum)
	return hsh.Sum(nil), nil
}

// Sum produces the public input
// Sum ignores NbBatches, as its value is expected to be incorporated into batchesSum.
func (i *FunctionalPublicInputQSnark) Sum(api frontend.API, hsh snarkHash.FieldHasher, batchesSum frontend.Variable) frontend.Variable {
	radix := big.NewInt(256)
	hsh.Reset()
	hsh.Write(
		compress.ReadNum(api, i.X[:16], radix),
		compress.ReadNum(api, i.X[16:], radix),
		i.Y[0],
		api.Add(api.Mul(i.Y[1], 256), i.Eip4844Enabled),
		i.SnarkHash,
		batchesSum,
	)
	return hsh.Sum()
}

// Sum hashes the inputs together into a single "de facto" public input
// WARNING: i.X[-] are not range-checked here
func (i *FunctionalPublicInputSnark) Sum(api frontend.API) (sum frontend.Variable, batchEvaluationPoints []frontend.Variable) {
	var (
		hsh        snarkHash.FieldHasher
		compressor snarkHash.Compressor
		err        error
	)

	if hsh, err = gkrposeidon2.New(api); err != nil {
		panic(err)
	}
	if compressor, err = poseidon2permutation.NewCompressor(api); err != nil {
		panic(err)
	}

	batchesToHash := make([]frontend.Variable, 2*len(i.BatchSumsNative))
	batchEvaluationPoints = make([]frontend.Variable, len(i.BatchSumsNative))
	for n := range i.BatchSumsNative {
		batchEvaluationPoints[n] = compressor.Compress(i.BatchSumsNative[n], i.BatchSumsSmallField[n])
		batchesToHash[2*n] = batchEvaluationPoints[n]
		batchesToHash[2*n+1] = i.BatchSumsTotal[n]
	}
	allBatchesHash := snarkHash.SumMerkleDamgardDynamicLength(api, compressor, 0, api.Mul(2, i.NbBatches), batchesToHash)

	return i.FunctionalPublicInputQSnark.Sum(api, hsh, allBatchesHash), batchEvaluationPoints
}

func (c Circuit) Define(api frontend.API) error {

	if c.CircuitSizes.DictNbBytes != len(c.Dict) {
		return fmt.Errorf("dictionary length mismatch: expected %d, got %d", c.CircuitSizes.DictNbBytes, len(c.Dict))
	}

	var (
		err        error
		hsh        snarkHash.StateStorer
		compressor snarkHash.Compressor
	)
	if hsh, err = gkrposeidon2.New(api); err != nil {
		return err
	}
	if compressor, err = poseidon2permutation.NewCompressor(api); err != nil {
		return err
	}

	// validate the input's form and range
	if err = c.FuncPI.Check(api, c.CircuitSizes); err != nil {
		return err
	}

	blobCrumbs := internal.PackedBytesToCrumbs(api, c.BlobBytes, blob.PackingSizeU256)

	blobPacked377 := internal.PackFull(api, blobCrumbs, 2) // repack into bls12-377 elements to compute a checksum
	hsh.Reset()
	hsh.Write(blobPacked377...)
	api.AssertIsEqual(c.FuncPI.SnarkHash, hsh.Sum())

	// EIP-4844 stuff
	if evaluation, err := publicinput.VerifyBlobConsistency(api, blobCrumbs, c.FuncPI.X, c.FuncPI.Eip4844Enabled); err != nil {
		return err
	} else {
		api.AssertIsEqual(c.FuncPI.Y[0], evaluation[0])
		api.AssertIsEqual(c.FuncPI.Y[1], evaluation[1])
	}

	// repack into bytes TODO possible optimization: pass bits directly to decompressor
	// unpack into bytes
	blobUnpackedBytes, blobUnpackedNbBytes := crumbStreamToByteStream(api, blobCrumbs)

	// get header length, number of batches, and length of each batch
	headerLen, dictChecksum, nbBatches, bytesPerBatch, err := parseHeader(api, c.MaxNbBatches, blobUnpackedBytes[:maxBlobNbBytes], blobUnpackedNbBytes)
	if err != nil {
		return err
	}
	api.AssertIsEqual(nbBatches, c.FuncPI.NbBatches)

	// check if the given decompression dictionary checksum matches the one used.
	if err = CheckDictChecksum(api, dictChecksum, c.Dict); err != nil {
		return err
	}

	// decompress the batches
	payload := make([]frontend.Variable, c.MaxUncompressedNbBytes)
	payloadLen, err := lzss.Decompress(
		api,
		compress.ShiftLeft(api, blobUnpackedBytes[:maxBlobNbBytes], headerLen), // TODO Signal to the decompressor that the input is zero padded; to reduce constraint numbers
		api.Sub(blobUnpackedNbBytes, headerLen),
		payload,
		c.Dict,
	)
	if err != nil {
		return err
	}
	api.AssertIsDifferent(payloadLen, -1) // decompression should not fail

	publicInput, batchEvaluationPoints := c.FuncPI.Sum(api)
	api.AssertIsEqual(c.PublicInput, publicInput)

	// validate "total" batch hashes
	batchEvaluations := [2][]frontend.Variable{c.FuncPI.BatchSumsNative, c.FuncPI.BatchSumsTotal}

	// compute checksum for each batch
	if err = CheckBatchesSums(api, compressor, c.MaxNbBatches, nbBatches, payload, bytesPerBatch, batchEvaluationPoints, batchEvaluations); err != nil {
		return err
	}

	return nil
}

type builder struct {
	config.CircuitSizes
}

func NewBuilder(cfg config.CircuitSizes) *builder {
	return &builder{cfg}
}

// Compile the decompression circuit
// Make sure to add the gkrmimc solver options in proving time
func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return Compile(b.CircuitSizes)
}

func Compile(sizes config.CircuitSizes) (constraint.ConstraintSystem, error) {
	return frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &Circuit{
		Dict:         make([]frontend.Variable, sizes.DictNbBytes),
		BlobBytes:    make([]frontend.Variable, blob.MaxUsableBytes),
		CircuitSizes: sizes,
	}, frontend.WithCapacity(1<<27))
}

func AssignFPI(blobBytes []byte, dictStore dictionary.Store, eip4844Enabled bool, x [32]byte, y fr381.Element) (fpi FunctionalPublicInput, dict []byte, err error) {
	if len(blobBytes) != blob.MaxUsableBytes {
		err = fmt.Errorf("decompression circuit assignment : invalid blob length : %d. expected %d", len(blobBytes), blob.MaxUsableBytes)
		return
	}

	r, err := blob.DecompressBlob(blobBytes, dictStore)
	if err != nil {
		return
	}
	dict = r.Dict

	if len(r.RawPayload) > blob.MaxUncompressedBytes {
		err = fmt.Errorf("decompression circuit assignment: blob payload too large : %d. max %d", len(r.RawPayload), blob.MaxUncompressedBytes)
		return
	}

	batchEnds := make([]int, r.Header.NbBatches())
	if r.Header.NbBatches() > 0 {
		batchEnds[0] = r.Header.BatchSizes[0]
	}
	for i := 1; i < len(r.Header.BatchSizes); i++ {
		batchEnds[i] = batchEnds[i-1] + r.Header.BatchSizes[i]
	}

	if fpi.BatchSums, err = batchesChecksumAssign(batchEnds, r.RawPayload); err != nil {
		return
	}

	fpi.X = x

	fpi.Y, err = internal.Bls12381ScalarToBls12377Scalars(y)
	if err != nil {
		return
	}

	fpi.Eip4844Enabled = eip4844Enabled

	if len(blobBytes) != 128*1024 {
		panic("blobBytes length is not 128*1024")
	}
	fpi.SnarkHash, err = encode.Poseidon2ChecksumPackedData(blobBytes, fr381.Bits-1, encode.NoTerminalSymbol()) // TODO if forced to remove the above check, pad with zeros

	return
}

func Assign(circuitSizes config.CircuitSizes, blobBytes []byte, dictStore dictionary.Store, eip4844Enabled bool, x [32]byte, y fr381.Element) (assignment frontend.Circuit, publicInput fr377.Element, snarkHash []byte, err error) {
	registerHints(circuitSizes.MaxNbBatches)
	fpi, dict, err := AssignFPI(blobBytes, dictStore, eip4844Enabled, x, y)
	if err != nil {
		return
	}
	snarkHash = fpi.SnarkHash

	pi, err := fpi.Sum()
	if err != nil {
		return
	}
	if err = publicInput.SetBytesCanonical(pi); err != nil {
		return
	}

	sfpi, err := fpi.ToSnarkType(circuitSizes.MaxNbBatches)
	if err != nil {
		return
	}

	assignment = &Circuit{
		Dict:        utils.ToVariableSlice(dict),
		BlobBytes:   utils.ToVariableSlice(blobBytes),
		PublicInput: publicInput,
		FuncPI:      sfpi,
	}

	return
}

// the result for each batch is <data (31 bytes)> ... <data (31 bytes)>
// for soundness some kind of length indicator must later be incorporated.
func batchesChecksumAssign(ends []int, payload []byte) ([]BatchSums, error) {

	res := make([]BatchSums, len(ends))

	nativeHash := gcHash.POSEIDON2_BLS12_377.New()
	smallFieldHash := gcHash.POSEIDON2_KOALABEAR.New()

	var batchNativeBuffer bytes.Buffer

	batchStart := 0
	for i := range res {
		nativeHash.Reset()
		smallFieldHash.Reset()

		if err := gnarkutil.PackLoose(smallFieldHash, payload[batchStart:ends[i]], koalabear.Bytes, smallFieldHash.BlockSize()/koalabear.Bytes); err != nil {
			return nil, err
		}

		batchNativeBuffer.Reset()
		if err := gnarkutil.PackLoose(&batchNativeBuffer, payload[batchStart:ends[i]], fr377.Bytes, 1); err != nil {
			return nil, err
		}

		if _, err := nativeHash.Write(batchNativeBuffer.Bytes()); err != nil {
			return nil, err
		}

		res[i].Native = nativeHash.Sum(nil)
		res[i].SmallField = smallFieldHash.Sum(nil)

		evaluationPoint, err := res[i].EvaluationPoint()
		if err != nil {
			return nil, err
		}

		if res[i].Total, err = polyEvalBls12377(batchNativeBuffer.Bytes(), evaluationPoint); err != nil {
			return nil, err
		}

		batchStart = ends[i]
	}

	return res, nil
}

func polyEvalBls12377(data []byte, evaluationPoint []byte) ([]byte, error) {
	var res, c, x fr377.Element
	if err := x.SetBytesCanonical(evaluationPoint); err != nil {
		return nil, err
	}
	nbBlocks := len(data) / fr377.Bytes
	if nbBlocks*fr377.Bytes != nbBlocks {
		return nil, fmt.Errorf("data must consist of %d-byte blocks, but the length is %d bytes", fr377.Bytes, len(data))
	}
	for i := range nbBlocks {
		if err := c.SetBytesCanonical(data[i*fr377.Bytes : (i+1)*fr377.Bytes]); err != nil {
			return nil, err
		}
		res.Mul(&c, &x)
		res.Add(&res, &c)
	}
	return res.Marshal(), nil
}
