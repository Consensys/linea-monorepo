package fetchers_arithmetization

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	arith "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

const (
	// ChainIDOffset is the corresponding offset position for the chain ID
	// since it is a shift, -1 means no offset.
	ChainIDOffset = -3
)

// ChainIDFetcher is a struct used to fetch the chainIDs from the arithmetization's BlockDataCols
type ChainIDFetcher struct {
	// chainID
	ChainID       ifaces.Column // a column used to fetch the ChainID. The implementation is currently unaligned with respect to the number of limbs.
	NBytesChainID ifaces.Column // a column used to fetch the number of bytes of the ChainID limb data
}

// NewTimestampFetcher returns a new TimestampFetcher with initialized columns that are not constrained.
func NewChainIDFetcher(comp *wizard.CompiledIOP, name string, bdc *arith.BlockDataCols) ChainIDFetcher {
	size := bdc.Ct.Size()
	res := ChainIDFetcher{
		ChainID:       util.CreateCol(name, "CHAIN_ID", size, comp),
		NBytesChainID: verifiercol.NewConstantCol(field.NewElement(2), size, "N_BYTES_CHAIN_ID"), // 2 bytes for chainID
	}

	return res
}

// DefineTimestampFetcher specifies the constraints of the TimestampFetcher with respect to the BlockDataCols
func DefineChainIDFetcher(comp *wizard.CompiledIOP, fetcher *ChainIDFetcher, name string, bdc *arith.BlockDataCols) {
	// constrain fetcher ChainID to contain the value of the last block's chainID
	// we only populate the first entry of the ChainID column
	comp.InsertLocal(
		0,
		ifaces.QueryIDf("%s_%s", name, "LAST_LOCAL"),
		sym.Sub(
			fetcher.ChainID,                         // first position of the ChainID column
			column.Shift(bdc.DataLo, ChainIDOffset), // corresponding position in the arithmetization's BlockDataCols
		),
	)
}

// AssignTimestampFetcher assigns the data in the TimestampFetcher using data fetched from the BlockDataCols
func AssignChainIDFetcher(run *wizard.ProverRuntime, fetcher *ChainIDFetcher, bdc *arith.BlockDataCols) {
	size := bdc.Ct.Size()
	var (
		chainID = make([]field.Element, size)
	)
	fetchedChainID := bdc.DataLo.GetColAssignmentAt(run, size+ChainIDOffset)
	chainID[0].Set(&fetchedChainID)
	run.AssignColumn(fetcher.ChainID.GetColID(), smartvectors.NewConstant(chainID[0], size))
}
