package generic

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

var (

	// UnspecifiedHashingUsecase is the zero value for the packing use-case and
	// it should not be used unless for testing that the caller of some function
	// does not pass it.
	UnspecifiedHashingUsecase = HashingUsecase{}

	// MiMCUsecase represents using MiMC with the Miyaguchi-Preneel construction
	// over a single field.
	/* MiMCUsecase = HashingUsecase{
		PaddingStrat:       zeroPadding,
		LaneSizeBytes_:     31,
		NbOfLanesPerBlock_: 1,
	}*/
	// MiMCUsecase points to Poseidon2’s settings for now. If you really need MiMC, just change the var temporarily.
	MiMCUsecase = newDeprecatedMiMCUsecase()

	// KeccakUsecase represents using the Keccak hash function as in Ethereum
	KeccakUsecase = HashingUsecase{
		PaddingStrat:       keccakPadding,
		LaneSizeBytes_:     8,
		NbOfLanesPerBlock_: 17,
	}

	// Sha2Usecase represents using the Sha2 hash function.
	Sha2Usecase = HashingUsecase{
		PaddingStrat:       sha2Padding,
		LaneSizeBytes_:     4,
		NbOfLanesPerBlock_: 16,
	}

	// Poseidon2Usecase represents using the Poseidon2 hash function.
	Poseidon2UseCase = HashingUsecase{
		PaddingStrat:       zeroPadding,
		LaneSizeBytes_:     poseidon2.BlockSize * field.Bytes,
		NbOfLanesPerBlock_: 1,
	}
)

func newDeprecatedMiMCUsecase() HashingUsecase {
	fmt.Println("[WARNING] MiMCUsecase (Padding) is deprecated — using Poseidon2UseCase." +
		" If you really need MiMC, change var [MiMCUsecase] temporarily to use MiMC settings.")
	return Poseidon2UseCase
}

type paddingStrat int

const (
	zeroPadding paddingStrat = iota
	keccakPadding
	sha2Padding
)

type HashingUsecase struct {
	PaddingStrat       paddingStrat
	LaneSizeBytes_     int
	NbOfLanesPerBlock_ int
}

func (huc HashingUsecase) BlockSizeBytes() int {
	return huc.LaneSizeBytes_ * huc.NbOfLanesPerBlock_
}

func (h HashingUsecase) LaneSizeBytes() int {
	return h.LaneSizeBytes_
}

func (h HashingUsecase) NbOfLanesPerBlock() int {
	return h.NbOfLanesPerBlock_
}
