package fetchers_arithmetization

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput/utilities"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestTimestampFetcher tests the fetching of the timestamp data
func TestTimestampFetcher(t *testing.T) {

	// initialize sample block data from a mock test data CSV file
	ctBlockData := utilities.InitializeCsv("../testdata/blockdata_mock.csv", t)
	var (
		bdc     *BlockDataCols
		fetcher TimestampFetcher
	)

	cmp := wizard.Compile(func(b *wizard.Builder) {
		// register sample arithmetization columns
		bdc = &BlockDataCols{
			RelBlock: ctBlockData.GetCommit(b, "REL_BLOCK"),
			Inst:     ctBlockData.GetCommit(b, "INST"),
			Ct:       ctBlockData.GetCommit(b, "CT"),
			DataHi:   ctBlockData.GetCommit(b, "DATA_HI"),
			DataLo:   ctBlockData.GetCommit(b, "DATA_LO"),
		}
		// create a new timestamp fetcher
		fetcher = NewTimestampFetcher(b.CompiledIOP, "TIMESTAMP_FETCHER_FROM_ARITH", bdc)
		// constrain the timestamp fetcher
		DefineTimestampFetcher(b.CompiledIOP, &fetcher, "TIMESTAMP_FETCHER_FROM_ARITH", bdc)
	}, dummy.Compile)
	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		// assign the CSV columns
		ctBlockData.Assign(
			run,
			"REL_BLOCK",
			"INST",
			"CT",
			"DATA_HI",
			"DATA_LO",
		)
		// assign the timestamp fetcher
		AssignTimestampFetcher(run, fetcher, bdc)
		// two simple sanity checks based on the mock test data
		assert.Equal(t, fetcher.First.GetColAssignmentAt(run, 0), field.NewElement(0xa))
		assert.Equal(t, fetcher.Last.GetColAssignmentAt(run, 0), field.NewElement(0xcd))
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
