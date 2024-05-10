package v1

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/hash/mimc"
	test_vector_utils "github.com/consensys/gnark/std/utils/test_vectors_utils"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/internal"
	blob "github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v1"
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
	// The dictionary used in the compression algorithm
	Dict []frontend.Variable

	// The uncompressed and compressed data corresponds to the data that is
	// made available on L1 when we submit a blob of transactions. The circuit
	// proves that these two correspond to the same data. The uncompressed
	// data is then passed to the EVM execution circuit.
	BlobBytes []frontend.Variable

	// The final public input. It is the hash of
	// 	- the hash of the uncompressed message.
	// 	- the hash of the compressed message. aka the SnarKHash on the contract
	// 	- the X and Y evaluation points claimed on the contract
	//  TODO - whether we are using Eip4844
	PublicInput frontend.Variable `gnark:",public"`

	// The X and Y evaluation point that are "claimed" by the contract. The
	// circuit will verify that the polynomial describing the compressed data
	// does evaluate to Y when evaluated at point X. This needed to ensure that
	// both the keccak hash of the compressed data (claimed on the contract) and
	// later the blob-hash (when we support EIP4844). TODO update this comment
	X              [2]frontend.Variable
	Y              [2]frontend.Variable
	SnarkHash      frontend.Variable
	Eip4844Enabled frontend.Variable
	BatchesSum     frontend.Variable
}

func (c Circuit) Define(api frontend.API) error {
	blobSum, y, batchesSum, err := ProcessBlob(api, c.BlobBytes, c.X, c.Eip4844Enabled, c.Dict)
	if err != nil {
		return err
	}
	api.AssertIsEqual(c.SnarkHash, blobSum)
	api.AssertIsEqual(c.Y[0], y[0])
	api.AssertIsEqual(c.Y[1], y[1])
	api.AssertIsEqual(c.BatchesSum, batchesSum)

	hsh, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	hsh.Write(c.X[0], c.X[1], c.Y[0], c.Y[1], c.SnarkHash, c.Eip4844Enabled, c.BatchesSum)
	api.AssertIsEqual(c.PublicInput, hsh.Sum())

	return nil
}

type builder struct {
	dictionaryLength int
}

func NewBuilder(dictionaryLength int) *builder {
	return &builder{dictionaryLength: dictionaryLength}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return Compile(b.dictionaryLength), nil
}

func Compile(dictionaryLength int) constraint.ConstraintSystem {
	// TODO @gbotrel make signature return error...
	if cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &Circuit{
		Dict:      make([]frontend.Variable, dictionaryLength),
		BlobBytes: make([]frontend.Variable, blob.MaxUsableBytes),
	}, frontend.WithCapacity(1<<27)); err != nil {
		panic(err)
	} else {
		return cs
	}
}

func Assign(blobBytes, dict []byte, eip4844Enabled bool, x, y fr381.Element) (assignment frontend.Circuit, publicInput fr377.Element, snarkHash []byte, err error) {
	const compressedMaxDataSize = blob.MaxUsableBytes
	const decompressedMaxDataSize = blob.MaxUncompressedBytes

	if len(blobBytes) != blob.MaxUsableBytes {
		err = fmt.Errorf("decompression circuit assignment : invalid blob length : %d. expected %d", len(blobBytes), blob.MaxUsableBytes)
		return
	}

	header, payload, _, err := blob.DecompressBlob(blobBytes, dict)
	if err != nil {
		return
	}

	if header.NbBatches() > MaxNbBatches {
		err = fmt.Errorf("decompression circuit assignment : too many batches in the header : %d. max %d", header.NbBatches(), MaxNbBatches)
		return
	}
	batchEnds := make([]int, header.NbBatches())
	if header.NbBatches() > 0 {
		batchEnds[0] = header.BatchSizes[0]
	}
	for i := 1; i < len(header.BatchSizes); i++ {
		batchEnds[i] = batchEnds[i-1] + header.BatchSizes[i]
	}
	hsh := hash.MIMC_BLS12_377.New()
	if header.NbBatches() > 255 {
		err = errors.New("more batches than 255 not currently supported")
		return
	} else {
		hsh.Write([]byte{byte(header.NbBatches())})
	}

	batchChecksums := checksumBatchesAssign(batchEnds, payload)
	for i := range batchChecksums {
		hsh.Write(batchChecksums[i])
	}
	var zero [fr377.Bytes]byte
	for i := len(batchChecksums); i < MaxNbBatches; i++ {
		hsh.Write(zero[:])
	}
	batchesSum := hsh.Sum(nil)

	x377, err := internal.Bls12381ScalarToBls12377Scalars(x)
	if err != nil {
		return
	}

	y377, err := internal.Bls12381ScalarToBls12377Scalars(y)
	if err != nil {
		return
	}

	var eip4844Enabled377 [fr377.Bytes]byte
	if eip4844Enabled {
		eip4844Enabled377[fr377.Bytes-1] = 1
	}

	if len(blobBytes) != 128*1024 {
		panic("blobBytes length is not 128*1024")
	}
	if snarkHash, err = blob.MiMCChecksumPackedData(blobBytes, fr381.Bits-1, blob.NoTerminalSymbol()); err != nil { // TODO if forced to remove the above check, pad with zeros
		return
	}
	hsh.Reset()
	hsh.Write(x377[0])
	hsh.Write(x377[1])
	hsh.Write(y377[0])
	hsh.Write(y377[1])
	hsh.Write(snarkHash)
	hsh.Write(eip4844Enabled377[:])
	hsh.Write(batchesSum)

	publicInputBytes := hsh.Sum(nil)
	err = publicInput.SetBytesCanonical(publicInputBytes[:])

	assignment = &Circuit{
		Dict:           test_vector_utils.ToVariableSlice(dict),
		BlobBytes:      test_vector_utils.ToVariableSlice(blobBytes),
		PublicInput:    publicInput,
		X:              [2]frontend.Variable{x377[0], x377[1]},
		Y:              [2]frontend.Variable{y377[0], y377[1]},
		SnarkHash:      snarkHash,
		Eip4844Enabled: eip4844Enabled377[:],
		BatchesSum:     batchesSum,
	}

	registerHints() // @Alexandre.Belling right place for this?

	return
}

// the result for each batch is <data (31 bytes)> ... <data (31 bytes)>
// for soundness some kind of length indicator must later be incorporated.
func checksumBatchesAssign(ends []int, payload []byte) [][]byte {
	res := make([][]byte, len(ends))

	in := make([]byte, len(payload)+31)
	copy(in, payload) // pad with 31 bytes to avoid out of range panic

	hsh := hash.MIMC_BLS12_377.New()

	batchStart := 0
	for i := range res {

		res[i] = in[batchStart : batchStart+31]
		nbWords := (ends[i] - batchStart + 30) / 31 // take as few 31-byte words as possible to cover the batch
		batchSumEnd := nbWords*31 + batchStart
		for j := batchStart + 31; j < batchSumEnd; j += 31 {
			hsh.Reset()
			hsh.Write(res[i])
			hsh.Write(in[j : j+31])
			res[i] = hsh.Sum(nil)
		}
		batchStart = ends[i]

	}

	return res
}
