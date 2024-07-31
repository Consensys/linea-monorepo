package testdata

import (
	"fmt"
	"math/rand"

	"github.com/consensys/zkevm-monorepo/prover/backend/files"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
)

// it receives columns hashNum and toHash and generates GenDataModule.
func GenerateAndAssignGenDataModule(run *wizard.ProverRuntime, gdm *generic.GenDataModule,
	hashNumInt, toHashInt []int, path ...string) {

	var (
		size    = gdm.Limb.Size()
		limbs   = make([]field.Element, size)
		nBytes  = make([]field.Element, size)
		toHash  = make([]field.Element, size)
		index   = make([]field.Element, size)
		hashNum = make([]field.Element, size)
		rng     = rand.New(rand.NewSource(68768))

		nByteCol   = common.NewVectorBuilder(gdm.NBytes)
		limbCol    = common.NewVectorBuilder(gdm.Limb)
		hashNumCol = common.NewVectorBuilder(gdm.HashNum)
		toHashCol  = common.NewVectorBuilder(gdm.TO_HASH)
		indexCol   = common.NewVectorBuilder(gdm.Index)
	)

	for i := range hashNumInt {

		if i == 0 {
			index[i] = field.Zero()
		} else if hashNumInt[i] != hashNumInt[i-1] {
			index[i] = field.Zero()
		} else if toHashInt[i] == 0 {
			index[i] = index[i-1]
		} else {
			index[i].Add(&index[i-1], new(field.Element).SetOne())
		}

		toHash[i] = field.NewElement(uint64(toHashInt[i]))
		hashNum[i] = field.NewElement(uint64(hashNumInt[i]))
		numBytesInt, numBytesF := randNBytes(rng)
		nBytes[i] = numBytesF
		limbs[i] = randLimbs(rng, numBytesInt)
	}

	limbCol.PushSliceF(limbs)
	nByteCol.PushSliceF(nBytes)
	hashNumCol.PushSliceF(hashNum)
	indexCol.PushSliceF(index)
	toHashCol.PushSliceF(toHash)

	limbCol.PadAndAssign(run)
	nByteCol.PadAndAssign(run)
	hashNumCol.PadAndAssign(run)
	indexCol.PadAndAssign(run)
	toHashCol.PadAndAssign(run)

	if len(path) > 0 {

		oF := files.MustOverwrite(path[0])
		fmt.Fprint(oF, "TO_HASH,HASH_NUM,INDEX,NBYTES,LIMBS\n")

		for i := range hashNumInt {
			fmt.Fprintf(oF, "%v,%v,%v,%v,0x%v\n",
				toHash[i].String(),
				hashNum[i].String(),
				index[i].String(),
				nBytes[i].String(),
				limbs[i].Text(16),
			)
		}

		oF.Close()
	}

}

func randNBytes(rng *rand.Rand) (int, field.Element) {

	// nBytesInt must be in 1..=16
	var (
		nBytesInt = rng.Int31n(16) + 1
		nBytesF   = field.NewElement(uint64(nBytesInt))
	)

	return int(nBytesInt), nBytesF
}

func randLimbs(rng *rand.Rand, nBytes int) field.Element {

	var (
		resBytes = make([]byte, 16)
		_, _     = rng.Read(resBytes[:nBytes])
		res      = new(field.Element).SetBytes(resBytes)
	)

	return *res
}

// CreateGenDataModule is used for testing, it commits to the [generic.GenDataModule] columns,
func CreateGenDataModule(
	comp *wizard.CompiledIOP,
	name string,
	size int,
) (gbm generic.GenDataModule) {
	createCol := common.CreateColFn(comp, name, size)
	gbm.HashNum = createCol("HASH_NUM")
	gbm.Index = createCol("INDEX")
	gbm.Limb = createCol("LIMBS")
	gbm.NBytes = createCol("NBYTES")
	gbm.TO_HASH = createCol("TO_HASH")
	return gbm
}
