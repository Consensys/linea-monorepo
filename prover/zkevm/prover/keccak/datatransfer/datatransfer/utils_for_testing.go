package datatransfer

import (
	"crypto/rand"
	mrand "math/rand"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/keccak/generic"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

type table struct {
	data         data
	info         info
	hasInfoTrace bool
}
type data struct {
	hashNum, index, nByte, lc, lx []int
	limb                          [][16]byte
}
type info struct {
	hashNum        []int
	hashHI, hashLO [][16]byte
}

// AssignGBMfromTable is used for testing.
// It assigns the Gbm (arithmetization columns relevant to keccak) from a random table.
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

	toHash := make([]field.Element, size)
	for i := range table.data.nByte {
		if table.data.nByte[i] != 0 {
			toHash[i] = field.One()
		}
	}
	run.AssignColumn(gbm.Data.TO_HASH.GetColID(), smartvectors.LeftZeroPadded(toHash, targetSize))

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
	// 	gbm.Data.TO_HASH = comp.InsertCommit(round, gbmDef.Data.TO_HASH, size)
	gbm.Data.TO_HASH = comp.InsertCommit(round, ifaces.ColIDf("TO_HASH"), size)

	return gbm
}

// tableForTest generates random gbm tables for the test
func tableForTest(size int) (t table) {
	numHash := size / 7
	// it fills  up DataTrace  and outputs the inputs for hashes
	msg := dataTrace(&t, numHash, size)
	// it fills up the InfoTrace
	infoTrace(&t, numHash, msg)
	// set hasInfoTrace to true
	t.hasInfoTrace = true

	return t
}

// It fills up the data trace of the table.
func dataTrace(t *table, numHash, size int) [][]byte {
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
	ctr := 0
	for k := 0; k < numHash; k++ {
		for j := range limbs[k] {
			t.data.hashNum[ctr+j] = k + 1
			t.data.index[ctr+j] = j
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
	// fill up LX, LC
	t.data.lc = make([]int, inLen)
	t.data.lx = make([]int, inLen)
	for i := 0; i < inLen; i++ {
		t.data.lx[i] = mrand.Intn(2) //nolint
		if t.data.nByte[i] == 0 {
			t.data.lc[i] = 0
		} else {
			t.data.lc[i] = 1
		}
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

	return msg
}

// It fills up the info trace of the table.
func infoTrace(t *table, numHash int, msg [][]byte) {
	out := t.info
	out.hashNum = make([]int, numHash)
	out.hashHI = make([][16]byte, numHash)
	out.hashLO = make([][16]byte, numHash)
	// sanity check
	if len(msg) != numHash {
		panic(" needs one message per hash")
	}

	for i := range out.hashNum {
		out.hashNum[i] = i

		// compute the hash for each msg
		h := sha3.NewLegacyKeccak256()
		h.Write(msg[i])
		outHash := h.Sum(nil)

		//assign Hash_HI and Hash_LOW
		if len(outHash) != 2*maxNByte {
			panic("can not cut the hash-output into Two Byte16")

		}
		copy(out.hashLO[i][:], outHash[:maxNByte])
		copy(out.hashHI[i][:], outHash[maxNByte:])
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
