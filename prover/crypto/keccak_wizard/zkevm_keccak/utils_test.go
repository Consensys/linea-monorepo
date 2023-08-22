package zkevm_keccak_test

import (
	"crypto/rand"
	mrand "math/rand"

	hp "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/hashing/hash_proof"
	zk "github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/zkevm_keccak"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

type table struct {
	zk.Tables
}

// it generates random zkEVM tables for the test
func TableForTest(size int) zk.Tables {
	var t table
	numHash := 10
	// it fills  up DataTrace  and outputs the inputs for hashes
	msg := t.DataTrace(numHash, size)
	// it fills up the InfoTrace
	t.InfoTrace(numHash, msg)
	// set hasInfoTrace to true
	t.HasInfoTrace = true

	return t.Tables
}
func (t *table) DataTrace(numHash, size int) []hp.Bytes {
	in := t.InputTable
	inLen := 0 // the total size of 'DataTrace=in'

	// choose the limbs for each hash
	// we set the limbs to less than LENGTH bytes and then pads them to get LENGTH byte (exactly like zkEVM)
	limbs := make([][]hp.Bytes, numHash)
	//at the same time build the hash inputs
	msg := make([]hp.Bytes, numHash)
	s := make([]int, numHash)
	for i := 0; i < numHash; i++ {
		// added +1 to prevent H-cases
		nlimb := mrand.Intn(size-(numHash-i-1)*7-inLen) + 1 //nolint
		if i == numHash-1 {
			nlimb = size - inLen
		}
		limbs[i] = make([]hp.Bytes, nlimb)
		s[i] = 0
		for j := range limbs[i] {
			limbs[i][j] = make(hp.Bytes, mrand.Intn(zk.LENGTH)) //nolint
			_, err := rand.Read(limbs[i][j])
			if err != nil {
				logrus.Fatalf("error while generating random bytes: %s", err)
			}
			s[i] += len(limbs[i][j])
		}
		inLen += nlimb
	}
	if inLen != size {
		panic("size of the table is not %v , size")
	}
	// fill up the table 'DataTrace'
	in.HashNum = make([]int, inLen)
	in.Index = make([]int, inLen)
	in.Limb = make([]zk.Byte16, inLen)
	in.Nbytes = make([]int, inLen)
	ctr := 0
	for k := 0; k < numHash; k++ {
		for j := range limbs[k] {
			in.HashNum[ctr+j] = k
			in.Index[ctr+j] = j
			in.Limb[ctr+j] = ToByte16(limbs[k][j])
			in.Nbytes[ctr+j] = len(limbs[k][j])
		}
		ctr += len(limbs[k])
	}
	if ctr != inLen {
		panic("the length of the table  is not consistent with HASH_NUM and LIMB")
	}

	if len(msg) != numHash {
		panic("needs one message per hash")
	}

	// get the inputs for the hashes
	for i := range msg {
		for j := range limbs[i] {
			msg[i] = append(msg[i], limbs[i][j]...)
		}
		if len(msg[i]) != s[i] {
			utils.Panic("message is not set to the right length, message length %v, what it  should be %v", len(msg[i]), s[i])
		}
	}

	t.InputTable = in

	return msg
}

func (t *table) InfoTrace(numHash int, msg []hp.Bytes) {
	out := t.OutputTable
	out.HashNum = make([]int, numHash)
	out.HashHI = make([]zk.Byte16, numHash)
	out.HashLO = make([]zk.Byte16, numHash)
	// sanity check
	if len(msg) != numHash {
		panic(" needs one message per hash")
	}

	for i := range out.HashNum {
		out.HashNum[i] = i

		// compute the hash for each msg
		h := sha3.NewLegacyKeccak256()
		h.Write(msg[i])
		outHash := h.Sum(nil)

		//assign Hash_HI and Hash_LOW
		if len(outHash) != 2*zk.LENGTH {
			panic("can not cut the hash-output into Two Byte16")

		}
		copy(out.HashLO[i][:], outHash[:zk.LENGTH])
		copy(out.HashHI[i][:], outHash[zk.LENGTH:])
	}
	t.OutputTable = out

}

// it extends a short slice to [LENGTH]bytes
func ToByte16(b hp.Bytes) zk.Byte16 {
	if len(b) > zk.LENGTH {
		utils.Panic("the length of input should not be greater than %v", zk.LENGTH)
	}
	n := zk.LENGTH - len(b)
	a := make([]byte, n)
	var c [zk.LENGTH]byte
	b = append(b, a...)
	copy(c[:], b)
	return c
}
