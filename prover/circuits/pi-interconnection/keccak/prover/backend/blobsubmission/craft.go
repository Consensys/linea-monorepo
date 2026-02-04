package blobsubmission

import (
	"encoding/base64"
	"hash"

	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"golang.org/x/crypto/sha3"
)

var b64 = base64.StdEncoding

// Prepare a response object by computing all the fields except for the proof.

// TODO @gbotrel this is not used? confirm with @Tabaie / @AlexandreBelling
// Computes the SNARK hash of a stream of byte. Returns the hex string. The hash
// can fail if the input stream does not have the right format.

// Returns an evaluation challenge point from a SNARK hash and a blob hash. The
// evaluation challenge is obtained as the hash of the SnarkHash and the keccak
// hash (or the blob hash once we go EIP4844) in that order. The digest is
// returned as a field element modulo the scalar field of the curve BLS12-381.
func evaluationChallenge(snarkHash, keccakHash []byte) (x []byte) {

	// Use the keccak hash
	h := sha3.NewLegacyKeccak256()
	h.Write(snarkHash)
	h.Write(keccakHash)
	d := h.Sum(nil)
	return d
}

// Cast the list of the field into a vector of field elements and performs a
// polynomial evaluation in the scalar field of BLS12-381. The bytes are split
// in chunks of 31 bytes representing each a field element in bigendian order.
// The last chunk is padded to the right with zeroes. The input x is taken as
// an array of 32 bytes because the smart-contract generating it will be using
// the result of the keccak directly. The modular reduction is implicitly done
// during the evaluation of the compressed data polynomial representation.

// schnarfParts wrap the arguments needed to create a new Shnarf by calling
// the NewSchnarf() function.
type Shnarf struct {
	OldShnarf, SnarkHash, NewStateRootHash []byte
	X                                      []byte
	Y                                      fr381.Element
	Hash                                   hash.Hash
}

// Compute returns the new shnarf given, the old shnarf, the snark hash, the new state
// root hash, the x and the y in that order. All over 32 bytes.
func (s *Shnarf) Compute() []byte {
	if s.Hash == nil {
		s.Hash = sha3.NewLegacyKeccak256()
	}
	yBytes := s.Y.Bytes()
	s.Hash.Reset()
	s.Hash.Write(s.OldShnarf)
	s.Hash.Write(s.SnarkHash)
	s.Hash.Write(s.NewStateRootHash)
	s.Hash.Write(s.X)
	s.Hash.Write(yBytes[:])

	return s.Hash.Sum(nil)
}
