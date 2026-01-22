package execution_data_collector

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	util "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/utilities"
)

// PoseidonPadderPacker is used to MiMC-hash the data in LogMessages. Using a zero initial stata,
// the data in L2L1 logs must be hashed as follows: msg1Hash, msg2Hash, and so on.
// The final value of the chained hash can be retrieved as hash[ctMax[any index]]
type PoseidonPadderPacker struct {
	InputIsActive ifaces.Column
	InterIsActive ifaces.Column
	InputData     ifaces.Column
	InterData     ifaces.Column
	OutputData    [common.NbLimbU128]ifaces.Column
	// OutputIsActive indicates which of the output data limbs are active
	// use OutputIsActive[0] for hashing, as it has more 1s and will include the zero padding
	OutputIsActive [common.NbLimbU128]ifaces.Column
}

// NewPoseidonPadderPacker returns a new PoseidonPadderPacker with initialized columns that are not constrained.
func NewPoseidonPadderPacker(comp *wizard.CompiledIOP, inputData, inputIsActive ifaces.Column, name string) PoseidonPadderPacker {
	var (
		res     PoseidonPadderPacker
		newSize int
	)
	res.InputData = inputData
	res.InputIsActive = inputIsActive

	if inputData.Size()%common.NbLimbU128 == 0 {
		newSize = inputData.Size()
	} else {
		newSize = ((inputData.Size() / common.NbLimbU128) + 1) * common.NbLimbU128
	}
	res.InterData = util.CreateCol(name, "INTER_DATA", newSize, comp)
	res.InterIsActive = util.CreateCol(name, "INTER_IS_ACTIVE", newSize, comp)

	for i := range res.OutputData {
		res.OutputData[i] = util.CreateCol(name, fmt.Sprintf("OUTPUT_DATA_%d", i), newSize/8, comp)
		res.OutputIsActive[i] = util.CreateCol(name, fmt.Sprintf("OUTPUT_IS_ACTIVE_%d", i), newSize/8, comp)
	}

	return res
}

// DefineHasher specifies the constraints of the PoseidonPadderPacker with respect to the ExtractedData fetched from the arithmetization
func DefinePoseidonPadderPacker(comp *wizard.CompiledIOP, ppp PoseidonPadderPacker, name string) {

	// Needed for the limitless prover to understand that the columns are
	// not just empty with just padding and suboptimal representation.
	pragmas.MarkRightPadded(ppp.InterData)
	pragmas.MarkRightPadded(ppp.InterIsActive)
	for i := range ppp.OutputData {
		pragmas.MarkRightPadded(ppp.OutputData[i])
	}

	comp.InsertProjection(
		ifaces.QueryIDf("%s_POSEIDON_PADDER_REGROUPER_PROJECTION", name),
		query.ProjectionInput{
			// the table with the data we fetch from the arithmetization's TxnData columns
			ColumnA: []ifaces.Column{ppp.InputData, ppp.InputIsActive},
			// the TxnData we extract sender addresses from, and which we will use to check for consistency
			ColumnB: []ifaces.Column{ppp.InterData, ppp.InterIsActive},
			FilterA: ppp.InputIsActive,
			// filter lights up on the arithmetization's TxnData rows that contain sender address data
			FilterB: ppp.InterIsActive})

	commonconstraints.MustBeBinary(comp, ppp.InterIsActive)

	comp.InsertGlobal(0, ifaces.QueryIDf("%s_%s", name, "POSEIDON_PADDER_REGROUPER_ZEROIZATIO_FILTER"),
		sym.Mul(
			sym.Sub(1,
				ppp.InterIsActive,
			),
			ppp.InterData, // when not active, the inter data must be 0
		),
	)

	// prepare the projection query to check that the OutputData is packed correctly
	multiaryFilterIntermediary := []ifaces.Column{verifiercol.NewConstantCol(field.One(), ppp.InterData.Size(), string(ifaces.ColIDf("POSEIDON_PADDER_PACKER_FILTER_FIRST_MULTIARY_%s", name)))}
	for i := 1; i < common.NbLimbU128; i += 1 {
		multiaryFilterIntermediary = append(multiaryFilterIntermediary,
			verifiercol.NewConstantCol(field.Zero(), ppp.InterData.Size(), string(ifaces.ColIDf("POSEIDON_PADDER_PACKER_SMALL_INTER_FILTER_%s_%d", name, i))),
		)
	}

	multiaryTableIntermediary := [][]ifaces.Column{[]ifaces.Column{ppp.InterData, ppp.InterIsActive}}
	for i := 1; i < common.NbLimbU128; i += 1 {
		multiaryTableIntermediary = append(multiaryTableIntermediary,
			[]ifaces.Column{
				verifiercol.NewConstantCol(field.Zero(), ppp.InterData.Size(), string(ifaces.ColIDf("POSEIDON_PADDER_PACKER_SMALL_INTER_COLUMN_%s_%d", name, i))),
				verifiercol.NewConstantCol(field.Zero(), ppp.InterData.Size(), string(ifaces.ColIDf("POSEIDON_PADDER_PACKER_SMALL_INTER_COLUMN_%s_%d", name, i))),
			})
	}

	multiaryFilterOutput := []ifaces.Column{}
	for i := 0; i < common.NbLimbU128; i += 1 {
		multiaryFilterOutput = append(multiaryFilterOutput,
			verifiercol.NewConstantCol(field.One(), ppp.InterData.Size()/common.NbLimbU128, string(ifaces.ColIDf("POSEIDON_PADDER_PACKER_SMALL_FILTER_%s_%d", name, i))),
		)
	}

	multiaryTableOutput := make([][]ifaces.Column, common.NbLimbU128)
	for i := 0; i < common.NbLimbU128; i += 1 {
		multiaryTableOutput[i] = []ifaces.Column{ppp.OutputData[i], ppp.OutputIsActive[i]}
	}

	/*	q := query.ProjectionMultiAryInput{
		ColumnsA: multiaryTableIntermediary,
		FiltersA: multiaryFilterIntermediary,
		ColumnsB: multiaryTableOutput,
		FiltersB: multiaryFilterOutput,
	}*/

	//comp.InsertProjection(ifaces.QueryIDf("%s_PROJ_POSEIDON_PADDER_PACKER", name), q)
}

// AssignHasher assigns the data in the PoseidonPadderPacker using the ExtractedData fetched from the arithmetization
func AssignPoseidonPadderPacker(run *wizard.ProverRuntime, ppp PoseidonPadderPacker) {
	interIsActive := make([]field.Element, ppp.InterData.Size())
	interData := make([]field.Element, ppp.InterData.Size())
	outputData := make([][]field.Element, common.NbLimbU128)
	outputIsActive := make([][]field.Element, common.NbLimbU128)
	for index := 0; index < common.NbLimbU128; index++ {
		outputData[index] = make([]field.Element, ppp.InterData.Size()/common.NbLimbU128)
		outputIsActive[index] = make([]field.Element, ppp.InterData.Size()/common.NbLimbU128)
	}

	for i := 0; i < ppp.InterData.Size(); i++ {
		var fetchedValue, fetchedFilter field.Element
		if i < ppp.InputData.Size() {
			fetchedValue = ppp.InputData.GetColAssignmentAt(run, i)
			fetchedFilter = ppp.InputIsActive.GetColAssignmentAt(run, i)
		} else {
			fetchedValue.SetZero()
			fetchedFilter.SetZero()
		}
		interIsActive[i].Set(&fetchedFilter)
		if fetchedFilter.IsOne() {
			interData[i].Set(&fetchedValue)
		}
		outputData[i%common.NbLimbU128][i/common.NbLimbU128].Set(&fetchedValue)
		outputIsActive[i%common.NbLimbU128][i/common.NbLimbU128].Set(&fetchedFilter)

	}

	interIsActiveSV := smartvectors.NewRegular(interIsActive)
	InterDataSV := smartvectors.NewRegular(interData)
	run.AssignColumn(ppp.InterData.GetColID(), InterDataSV)         // filter on LogColumns
	run.AssignColumn(ppp.InterIsActive.GetColID(), interIsActiveSV) // filter on fetched data
	for index := 0; index < common.NbLimbU128; index++ {
		run.AssignColumn(ppp.OutputData[index].GetColID(), smartvectors.NewRegular(outputData[index]))
		run.AssignColumn(ppp.OutputIsActive[index].GetColID(), smartvectors.NewRegular(outputIsActive[index]))
	}

}
