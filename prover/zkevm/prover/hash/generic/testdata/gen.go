package testdata

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

// SplitBytes splits the input slice into subarrays of the provided size.
func SplitBytes(input []byte, limbByteSize int) [][]byte {
	if len(input) == 0 {
		return [][]byte{}
	}

	var result [][]byte
	for i := 0; i < len(input); i += limbByteSize {
		end := i + limbByteSize
		if end > len(input) {
			end = len(input)
		}
		result = append(result, input[i:end])
	}
	return result
}

// it receives columns hashNum and toHash and generates GenDataModule.If flag is true it generate random nBytes between 1 and 16. Otherwise it sets nBytes to 16.
func GenerateAndAssignGenDataModule(run *wizard.ProverRuntime, gdm *generic.GenDataModule,
	hashNumInt, toHashInt []int, flag bool, path ...string) {

	var (
		rng        = rand.New(rand.NewChaCha8([32]byte{}))
		limbCols   = common.NewMultiVectorBuilder(gdm.Limbs)
		nByteCol   = common.NewVectorBuilder(gdm.NBytes)
		hashNumCol = common.NewVectorBuilder(gdm.HashNum)
		toHashCol  = common.NewVectorBuilder(gdm.ToHash)
		indexCol   = common.NewVectorBuilder(gdm.Index)
	)

	for i := range hashNumInt {

		if i == 0 {
			indexCol.PushInt(0)
		} else if hashNumInt[i] != hashNumInt[i-1] {
			indexCol.PushInt(0)
		} else if toHashInt[i] == 0 {
			indexCol.PushIncBy(0)
		} else {
			indexCol.PushIncBy(1)
		}

		toHashCol.PushInt(toHashInt[i])
		hashNumCol.PushInt(hashNumInt[i])
		var numBytesInt int
		if flag {
			numBytesInt = int(rng.Int32N(16) + 1)
			nByteCol.PushInt(numBytesInt)
		} else {
			numBytesInt = 16
			nByteCol.PushInt(16)
		}

		resBytes := make([]byte, 16)
		_, _ = utils.ReadPseudoRand(rng, resBytes[:numBytesInt])
		limbs := common.Bytes16ToLimbsLe(resBytes)
		limbCols.PushRow(limbs[:])

	}

	nByteCol.PadAndAssign(run)
	hashNumCol.PadAndAssign(run)
	indexCol.PadAndAssign(run)
	toHashCol.PadAndAssign(run)
	limbCols.PadAssignZero(run)

	if len(path) > 0 {

		oF := files.MustOverwrite(path[0])
		fmt.Fprint(oF, "TO_HASH,HASH_NUM,INDEX,NBYTES,LIMBS\n")

		for i := range hashNumInt {
			var limbsStr []string
			for _, l := range limbCols.T {
				limbsStr = append(limbsStr, fmt.Sprintf("0x%s", l.Slice()[i].Text(16)))
			}

			fmt.Fprintf(oF, "%v,%v,%v,%v,0x%v\n",
				toHashCol.Slice()[i].String(),
				hashNumCol.Slice()[i].String(),
				indexCol.Slice()[i].String(),
				nByteCol.Slice()[i].String(),
				strings.Join(limbsStr, ","),
			)
		}

		oF.Close()
	}

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
		hashHi      = make([]*common.VectorBuilder, len(gim.HashHi))
		hashLo      = make([]*common.VectorBuilder, len(gim.HashHi))
		isHashHiCol = common.NewVectorBuilder(gim.IsHashHi)
		isHashLoCol = common.NewVectorBuilder(gim.IsHashLo)
	)

	for i := 0; i < len(gim.HashHi); i++ {
		hashHi[i] = common.NewVectorBuilder(gim.HashHi[i])
		hashLo[i] = common.NewVectorBuilder(gim.HashLo[i])
	}

	streams := gdm.ScanStreams(run)
	var res [][32]byte
	for _, stream := range streams {
		res = append(res, keccak.Hash(stream))

	}
	ctrHi := 0
	ctrLo := 0
	for i := range isHashHi {
		if isHashHi[i] == 1 {
			bytes := SplitBytes(res[ctrHi][:16], 2)
			for j := 0; j < len(gim.HashHi); j++ {
				hashHi[j].PushBytes(bytes[j])
			}
			isHashHiCol.PushInt(1)
			ctrHi++
		} else {
			for j := 0; j < len(gim.HashHi); j++ {
				hashHi[j].PushInt(0)
			}
			isHashHiCol.PushInt(0)
		}

		if isHashLo[i] == 1 {
			bytes := SplitBytes(res[ctrLo][16:], 2)
			for j := 0; j < len(gim.HashLo); j++ {
				hashLo[j].PushBytes(bytes[j])
			}
			isHashLoCol.PushInt(1)
			ctrLo++
		} else {
			for j := 0; j < len(gim.HashLo); j++ {
				hashLo[j].PushInt(0)
			}
			isHashLoCol.PushInt(0)
		}
	}

	for i := 0; i < len(gim.HashHi); i++ {
		hashHi[i].PadAndAssign(run)
		hashLo[i].PadAndAssign(run)
	}
	isHashHiCol.PadAndAssign(run)
	isHashLoCol.PadAndAssign(run)
}
