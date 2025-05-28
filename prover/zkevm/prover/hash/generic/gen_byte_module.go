package generic

import (
	"bytes"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	// TotalLimbSize is the total size of a limb in bytes.
	TotalLimbSize = 16
)

// GenericByteModule encodes the limbs with a left alignment approach as
// specified by the arithmetization. It contains two set of columns,
// Data and Info. This module can be used for various concrete
// arithmetization modules that call the hash module.
type GenericByteModule struct {
	// Data module summarizing the information about the data to hash.
	Data GenDataModule
	// Info module contains the result of the hash.
	Info GenInfoModule
}

// GenDataModule collects the columns summarizing the information about the
// data to hash.
type GenDataModule struct {
	HashNum ifaces.Column   // identifier for the hash
	Index   ifaces.Column   // identifier for the current limb
	Limbs   []ifaces.Column // the content of the limbs to hash
	NBytes  ifaces.Column   // indicates the total number of bytes to use from limbs cols
	ToHash  ifaces.Column
}

// GenInfoModule collects the columns summarizing information about the result of the hash
type GenInfoModule struct {
	HashNum  ifaces.Column   // Identifier for the hash. Allows joining with the data module
	HashLo   []ifaces.Column // The Low part of the hash result
	HashHi   []ifaces.Column // The Hi part of the hash result
	IsHashLo ifaces.Column   // indicating the location of HashHi
	IsHashHi ifaces.Column   // indicting the location od HashLo
}

// ScanStreams scans the receiver GenDataModule's assignment and returns the list
// of the byte stream encoded in the assignment.
func (gdm *GenDataModule) ScanStreams(run *wizard.ProverRuntime) [][]byte {

	var (
		numRow      = gdm.Limbs[0].Size()
		numСols     = uint64(len(gdm.Limbs))
		index       = gdm.Index.GetColAssignment(run).IntoRegVecSaveAlloc()
		toHash      = gdm.ToHash.GetColAssignment(run).IntoRegVecSaveAlloc()
		hashNum     = gdm.HashNum.GetColAssignment(run).IntoRegVecSaveAlloc()
		nByte       = gdm.NBytes.GetColAssignment(run).IntoRegVecSaveAlloc()
		streams     = [][]byte(nil)
		buffer      = &bytes.Buffer{}
		currHashNum field.Element
	)

	maxNbBytesPerLimb := (TotalLimbSize + numСols - 1) / numСols

	// considering left-alignment approach, nbUnusedBytes is the number
	// of unused bytes of the limb represented by the field element
	nbUnusedBytes := field.Bytes - maxNbBytesPerLimb

	limbs := make([][]field.Element, numСols)
	for i := uint64(0); i < numСols; i++ {
		limbs[i] = gdm.Limbs[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	for row := 0; row < numRow; row++ {

		if toHash[row].IsZero() {
			continue
		}

		if index[row].IsZero() {
			if !currHashNum.IsZero() {
				streams = append(streams, buffer.Bytes())
				buffer = &bytes.Buffer{}
			}
			currHashNum = hashNum[row]
		}

		currNbBytes := nByte[row].Uint64()
		for col := uint64(0); col < numСols; col++ {
			currLimb := limbs[col][row].Bytes()

			if (col+1)*maxNbBytesPerLimb <= currNbBytes {
				buffer.Write(currLimb[nbUnusedBytes : nbUnusedBytes+maxNbBytesPerLimb])
				continue
			}

			nonZeroBytes := currNbBytes % maxNbBytesPerLimb
			buffer.Write(currLimb[nbUnusedBytes : nbUnusedBytes+nonZeroBytes])
			break
		}
	}

	streams = append(streams, buffer.Bytes())
	return streams
}
