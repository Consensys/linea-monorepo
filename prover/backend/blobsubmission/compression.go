package blobsubmission

// Output of the compression prover
type Request struct {

	// If true, eip4844. If false or not defined, legacy calldata
	Eip4844Enabled bool `json:"eip4844Enabled"`

	// The compressed data in base64 string
	CompressedData string `json:"compressedData"`

	// Parent data hash: the hash of the compressed data that were last
	// submitted and following which we are submitted CompressedData.
	DataParentHash string `json:"dataParentHash"`

	// Conflation order
	ConflationOrder ConflationOrder `json:"conflationOrder"`
	// The parent zkRootHash for the succession of blocks. In hexstring.
	ParentStateRootHash string `json:"parentStateRootHash"`
	// The new state root hash
	FinalStateRootHash string `json:"finalStateRootHash"`

	// The previous shnarf
	PrevShnarf string `json:"prevShnarf"`
}

type Response struct {

	// If true, eip4844. If false or not defined, legacy calldata
	Eip4844Enabled bool `json:"eip4844Enabled"`

	// BlobHash (VersionedHash): Hash of the compressed data
	DataHash string `json:"dataHash"`
	// Blob: compressed data in base64 string
	CompressedData string `json:"compressedData"` // kzg4844.Blob [131072]byte

	// The KZG commitment of the blob-data
	Commitment string `json:"commitment"` // kzg4844.Commitment [48]byte
	// The KZG proof for the blob data consistency check
	KzgProofContract string `json:"kzgProofContract"` // kzg4844.Proof [48]byte
	// The KZG proof for the blob sidecar in the blob tx
	KzgProofSidecar string `json:"kzgProofSidecar"` // kzg4844.Proof [48]byte

	// The expected value of X and Y from the prover's perspective. In hexstring
	// as a field element on the BLS12 field.
	ExpectedX string `json:"expectedX"` //ExpectedX kzg4844.Point [32]byte
	ExpectedY string `json:"expectedY"` //ExpectedY kzg4844.Claim [32]byte

	// The Snark friendly hash of the inputs
	SnarkHash string `json:"snarkHash"`

	// Conflation order
	ConflationOrder ConflationOrder `json:"conflationOrder"`
	// (parentZkRootHash) The parent zkRootHash for the succession of blocks. In hexstring.
	ParentStateRootHash string `json:"parentStateRootHash"`
	// (newStateRootHash) The last root hash after executing all the blocks in the blob.
	FinalStateRootHash string `json:"finalStateRootHash"`
	// Parent data hash. Namely, the hash of the blob of compressed data that
	// were last submitted.
	DataParentHash string `json:"parentDataHash"`
	// The expected value of the shnarf (or super-hash) that we expect the
	// contract to recover. In hexstring.
	ExpectedShnarf string `json:"expectedShnarf"`
	// The shnarf upon which we are towering the current blob.
	PrevShnarf string `json:"prevShnarf"`
}

type ConflationOrder struct {
	StartingBlockNumber int   `json:"startingBlockNumber"`
	UpperBoundaries     []int `json:"upperBoundaries"`
}

func (order *ConflationOrder) Range() (start, end int) {
	return order.StartingBlockNumber, order.UpperBoundaries[len(order.UpperBoundaries)-1]
}
