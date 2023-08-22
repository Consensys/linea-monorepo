/*
prover data interfaces: this is the code reading the hash data from zkEVM tables (given by Olivier)
which then would be passed to the keccak Multihash Prover
https://github.com/ConsenSys/zkevm-spec/issues/40#transaction-rlp
*/

package zkevm_keccak

import (
	"bytes"

	hp "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/hashing/hash_proof"

	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"golang.org/x/crypto/sha3"
)

const (
	// size of limbs from zkevm (this length is the same for all keccak related modules)
	LENGTH = 16
)

type Byte16 [LENGTH]byte

// inputs to the hashes,a table of four columns
type DataTrace struct {
	// number of hash
	HashNum []int `json:"TX_NUM"`
	// index for limbs of input
	Index []int `json:"INDEX"`
	// limbs of input (short limbs are padded by zero)
	Limb []Byte16 `json:"LIMB"`
	// real size of a limb
	Nbytes []int `json:"nBYTES"`
}

// outputs of the hashes
type InfoTrace struct {
	//a table of 3 columns
	HashNum []int    `json:"HASH_NUM"`
	HashHI  []Byte16 `json:"HASH_HI"`
	HashLO  []Byte16 `json:"HASH_LO"`
}
type Tables struct {
	InputTable  DataTrace
	OutputTable InfoTrace

	//for some modules we may not have InfoTrace
	HasInfoTrace bool
	//special columns for txRLP
	LX, LC []int
}
type Module uint8

const (
	TXRLP Module = iota
	// used for test over random tables
	RAND
	PhoneyRLP
)

// it extracts multiHash from the table
func (table Tables) MultiHashFromTable(numPerm int, module Module) (h hp.MultiHash) {

	in := table.InputTable
	out := table.OutputTable

	// edge-case the table is empty
	if len(in.HashNum) == 0 {
		return hp.MultiHash{
			InputHash:  []hp.Bytes{},
			OutputHash: []hp.Bytes{},
		}
	}

	// counter for the number of hashes
	newhash := 0
	// original limbs for each hash
	limbs := make([][]byte, 1)
	var t []byte
	last := len(in.HashNum) - 1
	for k := 0; k < last; k++ {
		if in.HashNum[k] == in.HashNum[k+1] {
			if (module == TXRLP && table.LX[k] == 1 && table.LC[k] == 1) || module != TXRLP {
				if in.Index[k+1] != in.Index[k]+1 {
					utils.Panic("inconsistency between INDEX and HASH_NUM,  HashNum(%v) =%v, equal HashNum(%v)=%v, but Index(%v)=%v and Index(%v)=%v are not succesive", k, in.HashNum[k], k+1, in.HashNum[k+1], k, in.Index[k], k+1, in.Index[k+1])
				}
				t = in.GetLimb(k)
				limbs[newhash] = append(limbs[newhash], t[:]...)
			}

		} else {
			if in.Index[k+1] != 0 {
				panic("the first index of a new hash should be zero")
			}
			if (module == TXRLP && table.LX[k] == 1 && table.LC[k] == 1) || module != TXRLP {
				t = in.GetLimb(k)
				limbs[newhash] = append(limbs[newhash], t[:]...)
			}

			newhash++
			limbs = append(limbs, []byte{})

		}

	}

	if len(in.HashNum) > 0 && (module == TXRLP && table.LX[last] == 1 && table.LC[last] == 1) || module != TXRLP {
		t = in.GetLimb(last)
		limbs[newhash] = append(limbs[newhash], t[:]...)

	}

	//sanity check
	if table.HasInfoTrace && len(out.HashNum) != len(limbs) {
		utils.Panic("the number of hashes from two tables DataTrace, InforTarce are not consistent")
	}

	for l := range limbs {
		// compute the hash for each hash input
		y := sha3.NewLegacyKeccak256()
		y.Write(limbs[l])
		outHash := y.Sum(nil)

		if table.HasInfoTrace {
			outTable := append(table.OutputTable.HashLO[l][:], table.OutputTable.HashHI[l][:]...)
			//sanity check
			if !bytes.Equal(outTable, outHash) {
				utils.Panic("the hash value from dataTrace and  InfoTrace are not consistent")
			}

		}

		h.InputHash = append(h.InputHash, limbs[l])
		h.OutputHash = append(h.OutputHash, outHash)
	}

	return h
}

func (in DataTrace) GetLimb(i int) []byte {
	//sanity check
	if len(in.Limb[i]) != LENGTH {
		utils.Panic("the LIMB should be of %v byte", LENGTH)
	}
	/*
		 to extract real limb from [LENGTH]byte LIMB, it remove the padded redundant zero (LSB part)
		and then concatenate the limbs to get the original message
	*/

	m := in.Limb[i]
	return m[:in.Nbytes[i]]
}
