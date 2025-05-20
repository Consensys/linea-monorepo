package testdata

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// SplitBytes splits the input slice into subarrays of the provided size.
func SplitBytes(input []byte, limbSize int) [][]byte {
	if len(input) == 0 {
		return [][]byte{}
	}

	var result [][]byte
	for i := 0; i < len(input); i += limbSize {
		end := i + limbSize
		if end > len(input) {
			end = len(input)
		}
		result = append(result, input[i:end])
	}
	return result
}

// it receives columns hashNum and toHash and generates GenDataModule.
func GenerateAndAssignGenDataModule(run *wizard.ProverRuntime, gdm *generic.GenDataModule,
	hashNumInt, toHashInt []int, flag bool, path ...string) {

	var (
		size    = gdm.Limbs[0].Size()
		nBytes  = make([]field.Element, size)
		toHash  = make([]field.Element, size)
		index   = make([]field.Element, size)
		hashNum = make([]field.Element, size)

		limbs    = make([][]field.Element, len(gdm.Limbs))
		limbCols = make([]*common.VectorBuilder, len(gdm.Limbs))

		rng = rand.New(rand.NewChaCha8([32]byte{}))

		nByteCol   = common.NewVectorBuilder(gdm.NBytes)
		hashNumCol = common.NewVectorBuilder(gdm.HashNum)
		toHashCol  = common.NewVectorBuilder(gdm.ToHash)
		indexCol   = common.NewVectorBuilder(gdm.Index)
	)

	for i := 0; i < len(gdm.Limbs); i++ {
		limbCols[i] = common.NewVectorBuilder(gdm.Limbs[i])
	}

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
		var numBytesInt int
		var numBytesF field.Element
		if flag {
			numBytesInt, numBytesF = randNBytes(rng)
			nBytes[i] = numBytesF
		} else {
			nBytes[i] = field.NewElement(16)
			numBytesInt = 16
		}

		randElement := randLimbs(rng, numBytesInt)
		limbBytes := randElement.Bytes()
		dividedLimbs := SplitBytes(limbBytes[16:], 16/len(gdm.Limbs))

		for j, limb := range dividedLimbs {
			var bytes [16]byte
			copy(bytes[:], limb)

			var l field.Element
			l.SetBytes(bytes[:])

			limbs[j] = append(limbs[j], l)
		}

	}

	nByteCol.PushSliceF(nBytes)
	hashNumCol.PushSliceF(hashNum)
	indexCol.PushSliceF(index)
	toHashCol.PushSliceF(toHash)

	for i, col := range limbCols {
		col.PushSliceF(limbs[i])
		col.PadAndAssign(run)
	}

	nByteCol.PadAndAssign(run)
	hashNumCol.PadAndAssign(run)
	indexCol.PadAndAssign(run)
	toHashCol.PadAndAssign(run)

	if len(path) > 0 {

		oF := files.MustOverwrite(path[0])
		fmt.Fprint(oF, "TO_HASH,HASH_NUM,INDEX,NBYTES,LIMBS\n")

		for i := range hashNumInt {
			var limbsStr []string
			for _, l := range limbs[i] {
				limbsStr = append(limbsStr, fmt.Sprintf("0x%s", l.Text(16)))
			}

			fmt.Fprintf(oF, "%v,%v,%v,%v,0x%v\n",
				toHash[i].String(),
				hashNum[i].String(),
				index[i].String(),
				nBytes[i].String(),
				strings.Join(limbsStr, ","),
			)
		}

		oF.Close()
	}

}

func randNBytes(rng *rand.Rand) (int, field.Element) {

	// nBytesInt must be in 1..=16
	var (
		nBytesInt = rng.Int32N(16) + 1
		nBytesF   = field.NewElement(uint64(nBytesInt))
	)

	return int(nBytesInt), nBytesF
}

func randLimbs(rng *rand.Rand, nBytes int) field.Element {

	var (
		resBytes = make([]byte, 16)
		_, _     = utils.ReadPseudoRand(rng, resBytes[:nBytes])
		res      = new(field.Element).SetBytes(resBytes)
	)

	return *res
}

// CreateGenDataModule is used for testing, it commits to the [generic.GenDataModule] columns,
func CreateGenDataModule(comp *wizard.CompiledIOP, name string, size int, nbLimbs int) (gbm generic.GenDataModule) {
	createCol := common.CreateColFn(comp, name, size, pragmas.RightPadded)
	gbm.HashNum = createCol("HASH_NUM")
	gbm.Index = createCol("INDEX")

	for i := 0; i < nbLimbs; i++ {
		gbm.Limbs = append(gbm.Limbs, createCol("LIMBS_%d", i))
	}

	gbm.NBytes = createCol("NBYTES")
	gbm.ToHash = createCol("TO_HASH")
	return gbm
}

// CreateGenInfoModule is used for testing, it commits to the [generic.GenInfoModule] columns,
func CreateGenInfoModule(comp *wizard.CompiledIOP, name string, size int, nbLimbs int) (gim generic.GenInfoModule) {
	createCol := common.CreateColFn(comp, name, size, pragmas.RightPadded)

	for i := 0; i < nbLimbs; i++ {
		gim.HashHi = append(gim.HashHi, createCol("HASH_HI_%d", i))
		gim.HashLo = append(gim.HashLo, createCol("HASH_LO_%d", i))
	}

	gim.IsHashHi = createCol("IS_HASH_HI")
	gim.IsHashLo = createCol("IS_HASH_LO")
	return gim
}

// it embeds  the expected hash (for the steam encoded inside gdm) inside gim columns.
func GenerateAndAssignGenInfoModule(
	run *wizard.ProverRuntime,
	gim *generic.GenInfoModule,
	gdm generic.GenDataModule,
	isHashHi, isHashLo []int,
) {

	var (
		hashHi      = common.NewVectorBuilder(gim.HashHi[0])
		hashLo      = common.NewVectorBuilder(gim.HashLo[0])
		isHashHiCol = common.NewVectorBuilder(gim.IsHashHi)
		isHashLoCol = common.NewVectorBuilder(gim.IsHashLo)
	)
	streams := gdm.ScanStreams(run)
	var res [][32]byte
	for _, stream := range streams {
		res = append(res, keccak.Hash(stream))

	}
	ctrHi := 0
	ctrLo := 0
	for i := range isHashHi {
		if isHashHi[i] == 1 {
			hashHi.PushHi(res[ctrHi])
			isHashHiCol.PushInt(1)
			ctrHi++
		} else {
			hashHi.PushInt(0)
			isHashHiCol.PushInt(0)
		}

		if isHashLo[i] == 1 {
			hashLo.PushLo(res[ctrLo])
			isHashLoCol.PushInt(1)
			ctrLo++
		} else {
			hashLo.PushInt(0)
			isHashLoCol.PushInt(0)
		}
	}

	hashHi.PadAndAssign(run)
	hashLo.PadAndAssign(run)
	isHashHiCol.PadAndAssign(run)
	isHashLoCol.PadAndAssign(run)
}
