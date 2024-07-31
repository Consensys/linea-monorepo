package generic

var (

	// UnspecifiedHashingUsecase is the zero value for the packing use-case and
	// it should not be used unless for testing that the caller of some function
	// does not pass it.
	UnspecifiedHashingUsecase = HashingUsecase{}

	// MiMCUsecase represents using MiMC with the Miyaguchi-Preneel construction
	// over a single field.
	MiMCUsecase = HashingUsecase{
		paddingStrat:      zeroPadding,
		laneSizeBytes:     31,
		nbOfLanesPerBlock: 1,
	}

	// KeccakUsecase represents using the Keccak hash function as in Ethereum
	KeccakUsecase = HashingUsecase{
		paddingStrat:      keccakPadding,
		laneSizeBytes:     8,
		nbOfLanesPerBlock: 17,
	}

	// Sha2Usecase represents using the Sha2 hash function.
	Sha2Usecase = HashingUsecase{
		paddingStrat:      sha2Padding,
		laneSizeBytes:     4,
		nbOfLanesPerBlock: 16,
	}
)

type paddingStrat int

const (
	zeroPadding paddingStrat = iota
	keccakPadding
	sha2Padding
)

type HashingUsecase struct {
	paddingStrat      paddingStrat
	laneSizeBytes     int
	nbOfLanesPerBlock int
}

func (huc HashingUsecase) BlockSizeBytes() int {
	return huc.laneSizeBytes * huc.nbOfLanesPerBlock
}

func (h HashingUsecase) LaneSizeBytes() int {
	return h.laneSizeBytes
}

func (h HashingUsecase) NbOfLanesPerBlock() int {
	return h.nbOfLanesPerBlock
}
