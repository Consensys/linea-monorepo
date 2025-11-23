package public_input

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	gcposeidon2permutation "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/consensys/gnark-crypto/field/koalabear"
	gchash "github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/icza/bitio"

	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// ExecDataChecksum consists of information enabling the computation
// of a common fingerprint for execution data by the different parts of
// the prover - native to different fields.
//
// @AlexandreBelling @Tabaie this solution is overcomplicated.
// It is best to just have both proofs compute the same hash.
// In the medium term, we should extract the data column from the wizard and
// compute the BLS12-377 hash in the Plonk "outer proof".
// In the long term, the Data Availability proof should also move to KoalaBear
// and compute the same hash natively.
type ExecDataChecksum struct {

	// Length of the execution data in bytes
	Length uint64

	// Bls12377PartialHash consists of hₙ₋₁ where hₙ₋₁ = H(hₙ₋₂, cₙ₋₁), ..., h₁ = H(h₀, c₁), h₀ = c₀ and
	// the cᵢ are the data packed into 31-byte BLS12-377 scalars in Big-Endian fashion.
	// H here is the Poseidon2 compression function for BLS12-377 scalars.
	// The last element is padded with less significant zeros as needed.
	// This sum is not collision resistant. To become so it needs the data length
	// incorporated into it as well.
	Bls12377PartialHash types.Bytes32

	// Bls12377Hash is a modified Poseidon2 hash of the execution data over
	// the BLS12-377 scalar field. It is computed as H(Bls12377PartialHash, Length).
	Bls12377Hash types.Bytes32

	// KoalaBearHash is a Poseidon2 hash of the execution data over the KoalaBear field.
	// The eight KoalaBear elements are concatenated bitwise so that the first byte
	// of this is always 0, and as a result it can be interpreted as a BLS12-377 scalar.
	KoalaBearHash types.Bytes32

	// PartialEvaluation is a Reed-Solomon fingerprint of the execution data.
	// It is computed as ∑ᵢ cᵢ.xⁿ⁻¹⁻ⁱ for 0 ≤ i < n.
	// x is the EvaluationPoint.
	// The cᵢ are the data packed into 31-byte BLS12-377 scalars in Big-Endian fashion.
	// The last element is padded with less significant zeros as needed.
	// As with Bls12377PartialHash, due to the zero padding of cₙ₋₁, this is not
	// collision-resistant as is, and needs the Length incorporated into it.
	PartialEvaluation types.Bytes32

	// EvaluationPoint is the Fiat-Shamir challenge for the Evaluation fingerprint. It is computed as
	// H(ExecDataChecksum.Bls12377Hash, ExecDataChecksum.KoalaBearHash)
	// where H is a single Poseidon2 compression over the BLS12-377 scalar field.
	EvaluationPoint types.Bytes32

	// Evaluation is a Reed-Solomon fingerprint of the execution data.
	// It is computed as ExecDataChecksum.Length + ExecDataChecksum.PartialEvaluation * EvaluationPoint
	Evaluation types.Bytes32

	// TotalChecksum is the BLS12-377 Poseidon2 hash of the EvaluationPoint and Evaluation together.
	TotalChecksum types.Bytes32
}

type Execution struct {
	L2MessageServiceAddr         types.EthAddress
	ChainID                      uint64
	InitialBlockTimestamp        uint64
	FinalStateRootHash           [32]byte
	FinalBlockNumber             uint64
	FinalBlockTimestamp          uint64
	LastRollingHashUpdate        [32]byte
	LastRollingHashUpdateNumber  uint64
	InitialRollingHashUpdate     [32]byte
	FirstRollingHashUpdateNumber uint64
	DataChecksum                 ExecDataChecksum
	L2MessageHashes              [][32]byte
	InitialStateRootHash         [32]byte
	InitialBlockNumber           uint64
}

func (pi *Execution) Sum() []byte {
	hsh := gchash.POSEIDON2_BLS12_377.New()

	for i := range pi.L2MessageHashes {
		hsh.Write(pi.L2MessageHashes[i][:16])
		hsh.Write(pi.L2MessageHashes[i][16:])
	}
	l2MessagesSum := hsh.Sum(nil)

	hsh.Reset()

	hsh.Write(pi.DataChecksum.TotalChecksum[:])

	hsh.Write(l2MessagesSum)
	hsh.Write(pi.FinalStateRootHash[:])

	writeNum(hsh, pi.FinalBlockNumber)
	writeNum(hsh, pi.FinalBlockTimestamp)
	hsh.Write(pi.LastRollingHashUpdate[:16])
	hsh.Write(pi.LastRollingHashUpdate[16:])
	writeNum(hsh, pi.LastRollingHashUpdateNumber)
	hsh.Write(pi.InitialStateRootHash[:])
	writeNum(hsh, pi.InitialBlockNumber)
	writeNum(hsh, pi.InitialBlockTimestamp)
	hsh.Write(pi.InitialRollingHashUpdate[:16])
	hsh.Write(pi.InitialRollingHashUpdate[16:])
	writeNum(hsh, pi.FirstRollingHashUpdateNumber)
	writeNum(hsh, pi.ChainID)
	hsh.Write(pi.L2MessageServiceAddr[:])

	return hsh.Sum(nil)

}

func (pi *Execution) SumAsField() field.Element {

	sumBytes := pi.Sum()
	sum := new(field.Element).SetBytes(sumBytes)

	return *sum
}

func writeNum(hsh hash.Hash, n uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], n)
	hsh.Write(b[:])
}

func NewExecDataChecksum(data []byte) (sums ExecDataChecksum, err error) {
	sums.Length = uint64(len(data))

	blsCompressor := gcposeidon2permutation.NewDefaultPermutation()
	koalaBearHash := gchash.POSEIDON2_KOALABEAR.New()

	if err = gnarkutil.PackLoose(koalaBearHash, data, koalabear.Bytes, koalaBearHash.BlockSize()/koalabear.Bytes); err != nil {
		return
	}
	if sums.KoalaBearHash, err = koalaBearToBls12377(koalaBearHash.Sum(nil)); err != nil {
		return
	}

	var blsBuf bytes.Buffer

	nbBlsElements := (len(data)-1)/(fr377.Bytes-1) + 1
	blsBuf.Grow(nbBlsElements * fr377.Bytes)

	if err = gnarkutil.PackLoose(&blsBuf, data, fr377.Bytes, 1); err != nil {
		return
	}

	if len(data) == 0 {
		return ExecDataChecksum{}, errors.New("this hashing scheme doesn't support empty data")
	}

	packed := blsBuf.Bytes()
	hsh := packed[:fr377.Bytes]
	for i := 1; i < nbBlsElements; i++ {
		if hsh, err = blsCompressor.Compress(hsh, packed[i*fr377.Bytes:i*fr377.Bytes+fr377.Bytes]); err != nil {
			return
		}
	}
	copy(sums.Bls12377PartialHash[:], hsh)

	var length fr377.Element
	length.SetUint64(sums.Length)
	if hsh, err = blsCompressor.Compress(hsh, length.Marshal()); err != nil {
		return
	}
	copy(sums.Bls12377Hash[:], hsh)

	if hsh, err = blsCompressor.Compress(sums.Bls12377Hash[:], sums.KoalaBearHash[:]); err != nil {
		return
	}
	copy(sums.EvaluationPoint[:], hsh)
	var evaluationPoint fr377.Element
	if err = evaluationPoint.SetBytesCanonical(sums.EvaluationPoint[:]); err != nil {
		return
	}

	eval, err := polyEvalBls12377(blsBuf.Bytes(), evaluationPoint)
	if err != nil {
		return sums, err
	}
	sums.PartialEvaluation = eval.Bytes()

	eval.Mul(&eval, &evaluationPoint)
	eval.Add(&eval, &length)
	sums.Evaluation = eval.Bytes()

	if hsh, err = blsCompressor.Compress(sums.EvaluationPoint[:], sums.Evaluation[:]); err != nil {
		return
	}

	copy(sums.TotalChecksum[:], hsh)

	return
}

// polyEvalBls377 treats data[len(data)-32:], data[len(data)-64:len(data)-32], ... as c₀, c₁, ...
// and evaluates c₀ + c₁x + c₂x² + ...
func polyEvalBls12377(data []byte, x fr377.Element) (fr377.Element, error) {
	var res, c fr377.Element
	nbBlocks := len(data) / fr377.Bytes
	if nbBlocks*fr377.Bytes != len(data) {
		return res, fmt.Errorf("data must consist of %d-byte blocks, but the length is %d bytes", fr377.Bytes, len(data))
	}

	if len(data) == 0 {
		return fr377.Element{}, nil
	}
	if err := res.SetBytesCanonical(data[:fr377.Bytes]); err != nil {
		return res, err
	}
	for i := 1; i < nbBlocks; i++ {
		res.Mul(&res, &x)
		if err := c.SetBytesCanonical(data[i*fr377.Bytes : (i+1)*fr377.Bytes]); err != nil {
			return res, err
		}
		res.Add(&res, &c)
	}
	return res, nil
}

// koalaBearToBls12377 converts eight KoalaBear field elements to a single BLS12-377 scalar.
func koalaBearToBls12377(in []byte) (types.Bytes32, error) {
	if len(in) != 32 {
		return types.Bytes32{}, fmt.Errorf("input length must be 32 bytes")
	}

	var out bytes.Buffer
	out.Grow(32)
	out.WriteByte(0) // last bit is always 0.

	writer := bitio.NewWriter(&out)

	for i := range 8 {
		u := binary.BigEndian.Uint32(in[4*i : 4*i+4])
		if u&0x80000000 != 0 {
			return types.Bytes32{}, fmt.Errorf("32-bit element at index %d; the koalabear modulus is only 31 bits long", i)
		}
		if err := writer.WriteBitsUnsafe(uint64(u), 31); err != nil {
			return types.Bytes32{}, err
		}
	}
	if err := writer.Close(); err != nil {
		return types.Bytes32{}, err
	}

	var res types.Bytes32
	copy(res[:], out.Bytes())
	return res, nil
}
