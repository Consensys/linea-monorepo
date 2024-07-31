package datatransfer

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

type table struct {
	data         data
	info         info
	hasInfoTrace bool
}
type data struct {
	hashNum, index, nByte, toHash []int
	limb                          [][16]byte
}
type info struct {
	hashNum            []int
	hashLo, hashHi     [][16]byte
	isHashLo, isHashHi []int
}

// AssignGBMfromTable is used for testing.
// It assigns the Gbm (arithmetization columns relevant to keccak) from a random table.
// witSize is the effective size of the module (with no padding)
// hashNum is the number of hashes that are extracted from such table.
// It is exported since we are using it for testing in different packages.
func AssignGBMfromTable(run *wizard.ProverRuntime, gbm *generic.GenericByteModule) {
	targetSize := gbm.Data.Limb.Size()

	// To support the edge cases, the assignment may not complete the column
	size := targetSize - targetSize/15
	table := &table{}
	*table = tableForTest(size)
	limb := table.data.limb
	u := make([]field.Element, size)

	for i := range limb {
		u[i].SetBytes(limb[i][:])

	}
	a := smartvectors.ForTest(table.data.hashNum...)
	aa := smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(a), targetSize)
	run.AssignColumn(gbm.Data.HashNum.GetColID(), aa)
	b := smartvectors.ForTest(table.data.index...)
	bb := smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(b), targetSize)
	run.AssignColumn(gbm.Data.Index.GetColID(), bb)
	c := smartvectors.ForTest(table.data.nByte...)
	cc := smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(c), targetSize)
	run.AssignColumn(gbm.Data.NBytes.GetColID(), cc)
	run.AssignColumn(gbm.Data.Limb.GetColID(), smartvectors.LeftZeroPadded(u, targetSize))

	d := smartvectors.ForTest(table.data.toHash...)
	dd := smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(d), targetSize)
	run.AssignColumn(gbm.Data.TO_HASH.GetColID(), dd)

	// assign Info trace
	if gbm.Info != (generic.GenInfoModule{}) {
		hashLo := table.info.hashLo
		hashHi := table.info.hashHi
		v := make([]field.Element, len(hashLo))
		w := make([]field.Element, len(hashHi))

		//reverse the order of bytes, to use SetByte() that is based on bigEndian
		var s, t [maxNByte]byte
		var hashLoLE, hashHiLE [][maxNByte]byte
		for i := range hashLo {
			for j := range hashLo[0] {
				s[j] = hashLo[i][maxNByte-1-j]
				t[j] = hashHi[i][maxNByte-1-j]
			}
			hashLoLE = append(hashLoLE, s)
			hashHiLE = append(hashHiLE, t)
		}

		for i := range hashLo {
			v[i].SetBytes(hashLoLE[i][:])
			w[i].SetBytes(hashHiLE[i][:])
		}

		run.AssignColumn(gbm.Info.HashLo.GetColID(), smartvectors.LeftZeroPadded(v, targetSize))
		run.AssignColumn(gbm.Info.HashHi.GetColID(), smartvectors.LeftZeroPadded(w, targetSize))

		if len(gbm.Info.HashNum.GetColID()) != 0 {
			t := smartvectors.ForTest(table.info.hashNum...)
			tt := smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(t), targetSize)
			run.AssignColumn(gbm.Info.HashNum.GetColID(), tt)
		}

		z := smartvectors.ForTest(table.info.isHashLo...)
		run.AssignColumn(gbm.Info.IsHashLo.GetColID(), smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(z), targetSize))
		run.AssignColumn(gbm.Info.IsHashHi.GetColID(), smartvectors.LeftZeroPadded(smartvectors.IntoRegVec(z), targetSize))
	}
}

// CommitGBM is used for testing, it commits to the gbm columns,
// i.e., the set of arithmetization columns relevant to keccak.
// It is exported since we are using it for testing in different packages.
func CommitGBM(
	comp *wizard.CompiledIOP,
	round int,
	gbmDef generic.GenericByteModuleDefinition,
	size int,
) (gbm generic.GenericByteModule) {
	gbm.Data.HashNum = comp.InsertCommit(round, gbmDef.Data.HashNum, size)
	gbm.Data.Index = comp.InsertCommit(round, gbmDef.Data.Index, size)
	gbm.Data.Limb = comp.InsertCommit(round, gbmDef.Data.Limb, size)
	gbm.Data.NBytes = comp.InsertCommit(round, gbmDef.Data.NBytes, size)

	if gbmDef.Data.HashNum != generic.PHONEY_RLP.Data.HashNum {
		gbm.Data.TO_HASH = comp.InsertCommit(round, gbmDef.Data.TO_HASH, size)
	} else {
		gbm.Data.TO_HASH = comp.InsertCommit(round, ifaces.ColIDf("To-Hash"), size)
	}

	if gbmDef.Info != (generic.InfoDef{}) {
		if len(gbmDef.Info.HashNum) != 0 {
			gbm.Info.HashNum = comp.InsertCommit(round, gbmDef.Info.HashNum, size)
		}
		gbm.Info.HashLo = comp.InsertCommit(round, gbmDef.Info.HashLo, size)
		gbm.Info.HashHi = comp.InsertCommit(round, gbmDef.Info.HashHi, size)
		gbm.Info.IsHashLo = comp.InsertCommit(round, gbmDef.Info.IsHashLo, size)
		gbm.Info.IsHashHi = comp.InsertCommit(round, gbmDef.Info.IsHashHi, size)
	}
	return gbm
}

// tableForTest generates random gbm tables for the test
func tableForTest(size int) (t table) {
	numHash := size / 7
	// it fills  up DataTrace  and outputs the inputs for hashes
	msg, chosens := dataTrace(&t, numHash, size)
	// it fills up the InfoTrace
	infoTrace(&t, numHash, msg, chosens)
	// set hasInfoTrace to true
	t.hasInfoTrace = true

	return t
}

// It fills up the data trace of the table.
func dataTrace(t *table, numHash, size int) ([][]byte, []int) {
	inLen := 0 // the total size of 'DataTrace'

	// choose the limbs for each hash
	// we set the limbs to less than LENGTH bytes and then pads them to get LENGTH byte (exactly like zkEVM)
	limbs := make([][][]byte, numHash)
	//at the same time build the hash inputs
	msg := make([][]byte, numHash)
	s := make([]int, numHash)
	for i := 0; i < numHash; i++ {
		// added +1 to prevent edge-cases
		nlimb := mrand.Intn(size-(numHash-i-1)*5-inLen) + 1 //nolint
		if i == numHash-1 {
			nlimb = size - inLen
		}
		limbs[i] = make([][]byte, nlimb)
		s[i] = 0
		for j := range limbs[i] {
			// for big tests
			limbs[i][j] = make([]byte, mrand.Intn(maxNByte)+1) //nolint
			// for small tests
			//limbs[i][j] = make([]byte, 1) //nolint
			_, err := rand.Read(limbs[i][j])
			if err != nil {
				logrus.Fatalf("error while generating random bytes: %s", err)
			}
			s[i] += len(limbs[i][j])
		}
		inLen += nlimb
	}
	if inLen != size {
		utils.Panic("size of the table  expected to be %v but it is  %v ", size, inLen)
	}
	// fill up the table 'DataTrace'
	t.data.hashNum = make([]int, inLen)
	t.data.index = make([]int, inLen)
	t.data.limb = make([][16]byte, inLen)
	t.data.nByte = make([]int, inLen)
	t.data.toHash = make([]int, inLen)

	ctr := 0
	var chosen []int
	for k := 0; k < numHash; k++ {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(2))
		bit := nBig.Uint64()
		if bit == 1 {
			chosen = append(chosen, k+1)
		}
		for j := range limbs[k] {
			t.data.hashNum[ctr+j] = k + 1
			t.data.index[ctr+j] = j
			t.data.toHash[ctr+j] = int(bit)
			t.data.limb[ctr+j] = toByte16(limbs[k][j])
			t.data.nByte[ctr+j] = len(limbs[k][j])
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

	return msg, chosen
}

// It fills up the info trace of the table.
func infoTrace(t *table, numHash int, msg [][]byte, chosen []int) {
	out := t.info
	out.hashNum = make([]int, numHash)
	out.hashLo = make([][16]byte, numHash)
	out.hashHi = make([][16]byte, numHash)
	out.isHashLo = make([]int, numHash)
	out.isHashHi = make([]int, numHash)
	// sanity check
	if len(msg) != numHash {
		panic(" needs one message per hash")
	}
	for i := range out.hashNum {
		out.hashNum[i] = i + 1

		// compute the hash for each msg
		h := sha3.NewLegacyKeccak256()
		h.Write(msg[i])
		outHash := h.Sum(nil)
		//assign Hash_HI and Hash_LOW
		if len(outHash) != 2*maxNByte {
			panic("can not cut the hash-output into Two Byte16")

		}
		copy(out.hashHi[i][:], outHash[:maxNByte])
		copy(out.hashLo[i][:], outHash[maxNByte:])

		for _, choose := range chosen {
			if out.hashNum[i] == choose {
				out.isHashLo[i] = 1
				out.isHashHi[i] = 1
			}
		}
	}
	t.info = out

}

// It extends a short slice to [16]bytes.
func toByte16(b []byte) [16]byte {
	if len(b) > maxNByte {
		utils.Panic("the length of input should not be greater than %v", maxNByte)
	}
	n := maxNByte - len(b)
	a := make([]byte, n)
	var c [maxNByte]byte
	b = append(b, a...)
	copy(c[:], b)
	return c
}

// It gets the block from the ImportAndPadd module and check it against the permutation trace.
func blocksFromIPadd(
	run *wizard.ProverRuntime,
	iPadd importAndPadd,
	blockSize uint64,
) [][]byte {
	witSize := smartvectors.Density(iPadd.nByte.GetColAssignment(run))
	eNByte := iPadd.nByte.GetColAssignment(run).IntoRegVecSaveAlloc()
	eCleanLimb := iPadd.cleanLimb.GetColAssignment(run).IntoRegVecSaveAlloc()
	isNewHash := iPadd.isNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()

	isEndOfBlock := make([]field.Element, witSize)
	ctr := 0
	s := eNByte[0].Uint64()

	var block [][]byte
	var stream []byte
	nbyte := eNByte[0].Uint64()
	limbBytes := eCleanLimb[0].Bytes()
	usefulBytes := limbBytes[32-nbyte:]
	stream = append(stream, usefulBytes[:nbyte]...)
	for j := 1; j < witSize; j++ {

		// sanity check
		if isNewHash[j] == field.One() && s != 0 {
			utils.Panic(" the last block should be complete before launching a new hash")
		}
		nbyte := eNByte[j].Uint64()
		s = s + nbyte
		limbBytes := eCleanLimb[j].Bytes()
		usefulBytes := limbBytes[32-nbyte:]
		if s > blockSize || s == blockSize {
			s = s - blockSize
			res := usefulBytes[:(nbyte - s)]
			newBlock := append(stream, res...)
			block = append(block, newBlock)
			stream = usefulBytes[(nbyte - s):nbyte]
			isEndOfBlock[j] = field.One()
			ctr++
		} else {
			stream = append(stream, usefulBytes[:nbyte]...)
		}

	}
	return block
}
