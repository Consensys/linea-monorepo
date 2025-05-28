package generic

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name        string
	streams     [][]byte
	nbLimbsCols int
	bytesPerRow int
}

func TestScanByteStream(t *testing.T) {
	testCases := []testCase{
		{
			name: "Single limbs column (16 bytes per row)",
			streams: [][]byte{
				mustDecodeHex("0x067845465329789679797Ffed67658"),
				mustDecodeHex("0xab477997edff98089860"),
			},
			nbLimbsCols: 1,
			bytesPerRow: 2,
		},
		{
			name: "2 limb columns (8 bytes per column)",
			streams: [][]byte{
				mustDecodeHex("0x00112233445566778899aabbccddeeff"),
			},
			nbLimbsCols: 2,
			bytesPerRow: TotalLimbSize,
		},
		{
			name: "4 limb columns (4 bytes per column)",
			streams: [][]byte{
				mustDecodeHex("0x00112233445566778899aabbccddeeff"),
			},
			nbLimbsCols: 4,
			bytesPerRow: TotalLimbSize,
		},
		{
			name: "8 limb columns (2 bytes per column)",
			streams: [][]byte{
				mustDecodeHex("0x00112233445566778899aabbccddeeff"),
			},
			nbLimbsCols: 8,
			bytesPerRow: TotalLimbSize,
		},
		{
			name: "16 limb columns (1 byte per column)",
			streams: [][]byte{
				mustDecodeHex("0x00112233445566778899aabbccddeeff"),
			},
			nbLimbsCols: 16,
			bytesPerRow: TotalLimbSize,
		},
		{
			name: "Mixed size streams with 4 columns",
			streams: [][]byte{
				mustDecodeHex("0x00112233445566778899aabbccddeeff"),   // 16 bytes - exactly one row
				mustDecodeHex("0x0011223344"),                         // 5 bytes - partial row
				mustDecodeHex("0x00112233445566778899aabbccddeeff00"), // 17 bytes - spans multiple rows
			},
			nbLimbsCols: 4,
			bytesPerRow: TotalLimbSize,
		},
		{
			name: "Uneven division with 3 columns",
			streams: [][]byte{
				mustDecodeHex("0x00112233445566778899aabbccddeeff"), // 16 bytes
			},
			nbLimbsCols: 3, // 16/3 = 5.33 bytes per column
			bytesPerRow: TotalLimbSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testScanByteStream)
	}
}

func (tc *testCase) testScanByteStream(t *testing.T) {
	var (
		gdm              GenDataModule
		recoveredStreams [][]byte
	)

	comp := wizard.Compile(func(build *wizard.Builder) {
		gdm = GenDataModule{
			HashNum: build.RegisterCommit("A", 32),
			Limbs:   make([]ifaces.Column, tc.nbLimbsCols),
			ToHash:  build.RegisterCommit("C", 32),
			NBytes:  build.RegisterCommit("D", 32),
			Index:   build.RegisterCommit("E", 32),
		}

		for i := 0; i < tc.nbLimbsCols; i++ {
			gdm.Limbs[i] = build.RegisterCommit(ifaces.ColIDf("B_%d", i), 32)
		}
	})

	_ = wizard.Prove(comp, func(run *wizard.ProverRuntime) {

		tc.assignGdbFromStream(run, &gdm)
		recoveredStreams = gdm.ScanStreams(run)
	})

	for i := range tc.streams {
		assert.Equalf(t,
			utils.HexEncodeToString(tc.streams[i]),
			utils.HexEncodeToString(recoveredStreams[i]),
			"position %v", i,
		)
	}
}

func (tc *testCase) assignGdbFromStream(run *wizard.ProverRuntime, gdm *GenDataModule) {
	nbCols := len(gdm.Limbs)

	hashNum := common.NewVectorBuilder(gdm.HashNum)
	nbBytes := common.NewVectorBuilder(gdm.NBytes)
	toHash := common.NewVectorBuilder(gdm.ToHash)
	index := common.NewVectorBuilder(gdm.Index)

	limbs := make([]*common.VectorBuilder, nbCols)
	for j := 0; j < nbCols; j++ {
		limbs[j] = common.NewVectorBuilder(gdm.Limbs[j])
	}

	maxBytesPerLimb := (TotalLimbSize + nbCols - 1) / nbCols
	nbUnusedBytes := field.Bytes - maxBytesPerLimb

	// Iterate over the streams in test case
	for hashID, currStream := range tc.streams {
		ctr := 0

		// Split the stream into rows
		for i := 0; i < len(currStream); {
			currRowBytes := utils.Min(len(currStream)-i, tc.bytesPerRow)

			// Split the current row into columns
			for col := 0; col < nbCols; col++ {
				var currLimb [field.Bytes]byte

				bytesCurrLimb := maxBytesPerLimb
				if (col+1)*maxBytesPerLimb > currRowBytes {
					bytesCurrLimb = currRowBytes % maxBytesPerLimb
				}

				if bytesCurrLimb > 0 && i < len(currStream) {
					copy(currLimb[nbUnusedBytes:], currStream[i:i+bytesCurrLimb])
					i += bytesCurrLimb
				}

				limbs[col].PushLo(currLimb)
			}

			hashNum.PushInt(hashID + 1)
			index.PushInt(ctr)
			nbBytes.PushInt(currRowBytes)
			toHash.PushOne()
			ctr++
		}
	}

	for _, limb := range limbs {
		limb.PadAndAssign(run, field.Zero())
	}
	hashNum.PadAndAssign(run, field.Zero())
	nbBytes.PadAndAssign(run, field.Zero())
	toHash.PadAndAssign(run, field.Zero())
	index.PadAndAssign(run)
}

func mustDecodeHex(s string) []byte {
	b, _ := utils.HexDecodeString(s)
	return b
}
