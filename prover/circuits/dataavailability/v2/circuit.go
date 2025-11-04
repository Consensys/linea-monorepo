package v2

import (
	"errors"
	"fmt"
	"hash"
	"math/big"

	"github.com/consensys/gnark/std/compress/lzss"
	gkrposeidon2 "github.com/consensys/gnark/std/hash/poseidon2/gkr-poseidon2"
	poseidon2permutation "github.com/consensys/gnark/std/permutation/poseidon2/gkr-poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/dataavailability/publicinput"
	"github.com/consensys/linea-monorepo/prover/config"
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

type Config struct {
	MaxUncompressedNbBytes int
	MaxNbBatches           int
}

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
	Config

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
type FunctionalPublicInput struct {
	X              [32]byte
	Y              [32]byte
	SnarkHash      []byte
	Eip4844Enabled bool
	BatchSums      []BatchSums // BatchSums checksums for batches of execution data - each corresponding to one exec proof
}

// Check checks that values are within range
func (i *FunctionalPublicInputQSnark) Check(api frontend.API, config Config) error {
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

func (i *FunctionalPublicInputSnark) Check(api frontend.API, config Config) error {
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
		return res, errors.New("batches do not fit in circuit")
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

	hsh.Reset()
	for n := range i.BatchSums {
		hsh.Write(i.BatchSums[n].Native)
		hsh.Write(i.BatchSums[n].SmallField)
		hsh.Write(i.BatchSums[n].Total)
	}
	batchesSum := hsh.Sum(nil)

	hsh.Reset()
	hsh.Write(i.X[:16])
	hsh.Write(i.X[16:])
	hsh.Write(i.Y[:16])
	hsh.Write(i.Y[16:])
	hsh.Write(i.SnarkHash)
	hsh.Write(utils.Ite(i.Eip4844Enabled, []byte{1}, []byte{0}))
	hsh.Write(batchesSum)
	return hsh.Sum(nil), nil
}

// Sum produces the public input
// Sum ignores NbBatches, as its value is expected to be incorporated into batchesSum.
func (i *FunctionalPublicInputQSnark) Sum(api frontend.API, hsh snarkHash.FieldHasher, batchesSum frontend.Variable) frontend.Variable {
	radix := big.NewInt(256)
	hsh.Reset()
	hsh.Write(compress.ReadNum(api, i.X[:16], radix), compress.ReadNum(api, i.X[16:], radix), i.Y[0], i.Y[1], i.SnarkHash, i.Eip4844Enabled, batchesSum)
	return hsh.Sum()
}

// Sum hashes the inputs together into a single "de facto" public input
// WARNING: i.X[-] are not range-checked here
func (i *FunctionalPublicInputSnark) Sum(api frontend.API) frontend.Variable {
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

	batchesToHash := make([]frontend.Variable, 3*len(i.BatchSumsNative))
	for n := range i.BatchSumsNative {
		batchesToHash[3*n] = i.BatchSumsNative[n]
		batchesToHash[3*n+1] = i.BatchSumsNative[n]
		batchesToHash[3*n+2] = i.BatchSumsSmallField[n]
	}
	allBatchesHash := snarkHash.SumMerkleDamgardDynamicLength(api, compressor, 0, api.Mul(3, i.NbBatches), batchesToHash)

	return i.FunctionalPublicInputQSnark.Sum(api, hsh, allBatchesHash)
}

func (c Circuit) Define(api frontend.API) error {

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
	if err = c.FuncPI.Check(api, c.Config); err != nil {
		return err
	}

	blobCrumbs := internal.PackedBytesToCrumbs(api, c.BlobBytes, blob.PackingSizeU256)

	blobPacked377 := internal.PackFull(api, blobCrumbs, 2) // repack into bls12-377 elements to compute a checksum
	hsh.Reset()
	hsh.Write(blobPacked377...)
	blobSum := hsh.Sum()

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
	headerLen, dictChecksum, nbBatches, bytesPerBatch, err := parseHeader(api, blobUnpackedBytes[:maxBlobNbBytes], blobUnpackedNbBytes)
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

	batchEvaluationPoints := make([]frontend.Variable, c.MaxNbBatches)
	// validate "total" batch hashes
	for i := range c.MaxNbBatches {
		batchEvaluationPoints[i] = compressor.Compress(c.FuncPI.BatchSumsNative[i], c.FuncPI.BatchSumsSmallField[i])
	}
	batchEvaluations := [2][]frontend.Variable{c.FuncPI.BatchSumsNative, c.FuncPI.BatchSumsTotal}

	// compute checksum for each batch
	if err = CheckBatchesSums(api, compressor, c.MaxNbBatches, nbBatches, payload, bytesPerBatch, batchEvaluationPoints, batchEvaluations); err != nil {
		return err
	}

	api.AssertIsEqual(c.FuncPI.SnarkHash, blobSum)

	api.AssertIsEqual(c.PublicInput, c.FuncPI.Sum(api, hsh))
	return nil
}

/*
 *
*  type builder struct {
	dictionaryLength int
}

 func NewBuilder(dictionaryLength int) *builder {
	return &builder{dictionaryLength: dictionaryLength}
}*/

// Compile the decompression circuit
// Make sure to add the gkrmimc solver options in proving time
func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return Compile(b.dictionaryLength), nil
}

func (config Config) Compile() (constraint.ConstraintSystem, error) {
	if cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &Circuit{
		Dict:                  make([]frontend.Variable, dictionaryLength),
		BlobBytes:             make([]frontend.Variable, blob.MaxUsableBytes),
		MaxBlobPayloadNbBytes: blob.MaxUncompressedBytes,
		UseGkrMiMC:            true,
	}, frontend.WithCapacity(1<<27)); err != nil {
		panic(err)
	} else {
		return cs
	}
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

	if r.Header.NbBatches() > MaxNbBatches {
		err = fmt.Errorf("decompression circuit assignment : too many batches in the header : %d. max %d", r.Header.NbBatches(), MaxNbBatches)
		return
	}

	batchEnds := make([]int, r.Header.NbBatches())
	if r.Header.NbBatches() > 0 {
		batchEnds[0] = r.Header.BatchSizes[0]
	}
	for i := 1; i < len(r.Header.BatchSizes); i++ {
		batchEnds[i] = batchEnds[i-1] + r.Header.BatchSizes[i]
	}

	fpi.BatchSums = batchesChecksumAssign(batchEnds, r.RawPayload)

	fpi.X = x

	fpi.Y, err = internal.Bls12381ScalarToBls12377Scalars(y)
	if err != nil {
		return
	}

	fpi.Eip4844Enabled = eip4844Enabled

	if len(blobBytes) != 128*1024 {
		panic("blobBytes length is not 128*1024")
	}
	fpi.SnarkHash, err = encode.MiMCChecksumPackedData(blobBytes, fr381.Bits-1, encode.NoTerminalSymbol()) // TODO if forced to remove the above check, pad with zeros

	return
}

func init() {
	registerHints()
}

func Assign(blobBytes []byte, dictStore dictionary.Store, eip4844Enabled bool, x [32]byte, y fr381.Element) (assignment frontend.Circuit, publicInput fr377.Element, snarkHash []byte, err error) {

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

	sfpi, err := fpi.ToSnarkType()
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
func batchesChecksumAssign(ends []int, payload []byte) []BatchSums {
	res := make([]BatchSums, len(ends))

	in := make([]byte, len(payload)+31)
	copy(in, payload) // pad with 31 bytes to avoid out of range panic

	batchStart := 0
	for i := range res {
		res[i] = gnarkutil.ChecksumMiMCLooselyPackedBytes(payload[batchStart:ends[i]])
		batchStart = ends[i]
	}

	return res
}
