package logs

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/ifaces"
)

// ExtractedData contains the data extracted from the arithmetization logs:
// L2L1 case: already Keccak-hashed messages, which will be hashed again using MiMC
// RollingHash case: either the message number stored in the Lo part
// or the RollingHash stored in both Hi/Lo
type ExtractedData struct {
	Hi, Lo        ifaces.Column
	FilterArith   ifaces.Column
	FilterFetched ifaces.Column
}
