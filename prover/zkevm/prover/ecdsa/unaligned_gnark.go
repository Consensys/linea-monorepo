package ecdsa

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"

	"github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	nbRowsPerPublicKey = 4

	// gnarkDataLeftAlignmentOffset is the number of bits that were offset in UnalignedGnarkData.GnarkData.
	gnarkDataLeftAlignmentOffset = (generic.TotalLimbSize - generic.TotalLimbSize/common.NbLimbU128) * 8
)

type UnalignedGnarkData struct {
	IsPublicKey         ifaces.Column
	GnarkIndex          ifaces.Column
	GnarkPublicKeyIndex ifaces.Column
	GnarkData           [common.NbLimbU128]ifaces.Column

	// auxiliary columns
	IsIndex0     ifaces.Column
	IsIndex0Act  wizard.ProverAction
	IsIndex5     ifaces.Column
	IsIndex5Act  wizard.ProverAction
	IsIndex4     ifaces.Column
	IsIndex4Act  wizard.ProverAction
	IsIndex13    ifaces.Column
	IsIndex13Act wizard.ProverAction

	IsEcrecoverAndFetching   ifaces.Column
	IsNotPublicKeyAndPushing ifaces.Column

	Size int
}

type unalignedGnarkDataSource struct {
	Source     ifaces.Column
	IsActive   ifaces.Column
	IsPushing  ifaces.Column
	IsFetching ifaces.Column
	IsData     ifaces.Column
	IsRes      ifaces.Column
	Limb       [common.NbLimbU128]ifaces.Column
	SuccessBit ifaces.Column
	TxHash     [common.NbLimbU256]ifaces.Column
}

// TxSignatureGetter is a function that is expected a signature for a transaction
// hash and/or a integer id. In production, the function returns the signature
// from the transaction id but uses the txHash as a sanity-check.
type TxSignatureGetter func(i int, txHash []byte) (r, s, v *big.Int, err error)

func newUnalignedGnarkData(comp *wizard.CompiledIOP, size int, src *unalignedGnarkDataSource) *UnalignedGnarkData {
	createCol := createColFn(comp, NAME_UNALIGNED_GNARKDATA, size)
	res := &UnalignedGnarkData{
		IsPublicKey:         createCol("IS_PUBLIC_KEY"),
		GnarkIndex:          createCol("GNARK_INDEX"),
		GnarkPublicKeyIndex: createCol("GNARK_PUBLIC_KEY_INDEX"),

		IsEcrecoverAndFetching:   createCol("IS_ECRECOVER_AND_FETCHING"),
		IsNotPublicKeyAndPushing: createCol("IS_NOT_PUBLIC_KEY_AND_PUSHING"),

		Size: size,
	}

	for i := 0; i < common.NbLimbU128; i++ {
		res.GnarkData[i] = createCol(fmt.Sprintf("GNARK_DATA_%d", i))
	}

	res.csDataIds(comp)
	res.csIndex(comp, src)
	res.csProjectionEcRecover(comp, src)
	res.csTxHash(comp, src)
	res.csTxEcRecoverBit(comp, src)

	// we do not have to constrain the public key as it will be done in the
	// address sub-module

	// we do not need to constrain in cases where is_fetching = 1 as later in
	// the projection to aligned gnark data we do not use the data.

	return res
}

func (d *UnalignedGnarkData) Assign(run *wizard.ProverRuntime, src *unalignedGnarkDataSource, txSigs TxSignatureGetter) {
	d.assignUnalignedGnarkData(run, src, txSigs)
	d.assignHelperColumns(run, src)
	// depends on the value in the gnarkIndex column, so needs to be run after
	d.IsIndex0Act.Run(run)
	d.IsIndex4Act.Run(run)
	d.IsIndex5Act.Run(run)
	d.IsIndex13Act.Run(run)
}

func (d *UnalignedGnarkData) assignUnalignedGnarkData(run *wizard.ProverRuntime, src *unalignedGnarkDataSource, txSigs TxSignatureGetter) {

	// copies data from the ecrecover part and txn part. Then it also computes
	// the public key values and stores them in the corresponding rows.
	var (
		sourceSource     = run.GetColumn(src.Source.GetColID())
		sourceIsActive   = run.GetColumn(src.IsActive.GetColID())
		sourceSuccessBit = run.GetColumn(src.SuccessBit.GetColID())
		sourceLimb       [common.NbLimbU128]ifaces.ColAssignment
		sourceTxHash     [common.NbLimbU256]ifaces.ColAssignment

		// prependZeroCount counts the number of zero bytes that are used by
		// the current frame to fetch the data from its source. Its value
		// depends on whether the current frame is for ecrecover or for txn
		// signature.
		prependZeroCount uint

		// txCounter counts the number of transactions that have been found so far
		// it is used to determine the transaction index for which to fetch
		// a signature
		txCounter = 0

		resIsPublicKey  = common.NewVectorBuilder(d.IsPublicKey)
		resGnarkIndex   = common.NewVectorBuilder(d.GnarkIndex)
		resGnarkPkIndex = common.NewVectorBuilder(d.GnarkPublicKeyIndex)
		resGnarkData    = common.NewMultiVectorBuilder(d.GnarkData[:])
	)

	printRows := func(head int, rows [][common.NbLimbU128]field.Element) {
		for i := range rows {
			fmt.Printf("%04d: row=%04d ", i, head+i)
			for j := range rows[i] {
				fmt.Printf("%04v ", rows[i][7-j].Text(16))
			}
			fmt.Printf("\n")
		}
	}

	recoverPk := func(h [32]byte, r, s, v *big.Int) (pkX, pkY *big.Int) {

		fmt.Printf("recovering pubk for h=%x r=%v s=%v v=%v\n", h, r, s, v)

		// compute the expected public key
		var pk ecdsa.PublicKey
		if !v.IsUint64() {
			utils.Panic("v is not a uint64, v %v", v.String())
		}
		err := pk.RecoverFrom(h[:], uint(v.Uint64()-27), r, s)
		if err != nil {
			utils.Panic("error recovering public: err=%v v=%v r=%v s=%v", err.Error(), v.Uint64()-27, r.String(), s.String())
		}

		fmt.Printf("gnark pk = %v %v\n", pk.A.X.Text(16), pk.A.Y.Text(16))

		pkX, pkY = new(big.Int), new(big.Int)
		pk.A.X.BigInt(pkX)
		pk.A.Y.BigInt(pkY)
		return pkX, pkY
	}

	for i := 0; i < common.NbLimbU128; i++ {
		sourceLimb[i] = run.GetColumn(src.Limb[i].GetColID())
		if sourceLimb[i].Len() != d.Size {
			panic("unexpected source limb length")
		}
	}

	for i := 0; i < common.NbLimbU256; i++ {
		sourceTxHash[i] = run.GetColumn(src.TxHash[i].GetColID())
		if sourceTxHash[i].Len() != d.Size {
			panic("unexpected source hash length")
		}
	}

	if sourceSource.Len() != d.Size || sourceIsActive.Len() != d.Size || sourceSuccessBit.Len() != d.Size {
		panic("unexpected source length")
	}

	// one iteration per ecdsa check. i is always the start of a frame in the
	// traces.
ecdsaLoop:
	for i := 0; i < d.Size; {

		var (
			isActive         = sourceIsActive.Get(i)
			source           = sourceSource.Get(i)
			dataForCurrEcdsa = make([][common.NbLimbU128]field.Element, nbRowsPerGnarkPushing)
			r, s, v          *big.Int
			h                [32]byte
			buff             [32]byte
			head             = i
		)

		switch {
		case isActive.IsOne() && source.Cmp(&SOURCE_ECRECOVER) == 0:

			prependZeroCount = nbRowsPerEcRecFetching

			// copy h, r, s, v
			//		4		5		6		7		8		9		10		11
			// 		h0, 	h1, 	r0, 	r1, 	s0, 	s1, 	v0, 	v1
			for k := 4; k < 12; k++ {
				copy(dataForCurrEcdsa[k][:], common.GetTableRow(i+k-4, sourceLimb[:]))
			}

			copy(h[:], common.HiLoLimbsLeToBytesBe(dataForCurrEcdsa[4][:], dataForCurrEcdsa[5][:]))
			v = common.LimbsLeToBigInt(dataForCurrEcdsa[6][:], dataForCurrEcdsa[7][:])
			r = common.LimbsLeToBigInt(dataForCurrEcdsa[8][:], dataForCurrEcdsa[9][:])
			s = common.LimbsLeToBigInt(dataForCurrEcdsa[10][:], dataForCurrEcdsa[11][:])

			if !v.IsUint64() {
				printRows(i, dataForCurrEcdsa)
				utils.Panic("v is not a uint64, v %v; r=%v s=%v", v.String(), r.String(), s.String())
			}

			// The success bit
			dataForCurrEcdsa[12] = [common.NbLimbU128]field.Element{
				sourceSuccessBit.Get(i), // implictly followed by 7 zeroes
			}

			// The ecrecover bit
			dataForCurrEcdsa[13] = [common.NbLimbU128]field.Element{
				field.One(), // implictly followed by 7 zeroes
			}

			i += NB_ECRECOVER_INPUTS

		case isActive.IsOne() && source.Cmp(&SOURCE_TX) == 0:

			prependZeroCount = nbRowsPerTxSignFetching

			var (
				txHashHi = common.GetTableRow(i, sourceTxHash[:8])
				txHashLo = common.GetTableRow(i, sourceTxHash[8:])
				txHash   = common.HiLoLimbsLeToBytesBe(txHashHi, txHashLo)
				sigErr   error
			)

			r, s, v, sigErr = txSigs(txCounter, txHash)

			if sigErr != nil {
				utils.Panic("error getting tx-signature err=%v, txNum=%v", sigErr, txCounter)
			}
			if !v.IsUint64() {
				utils.Panic("v is not a uint64, v=%v", v.String())
			}

			copy(h[:], txHash)

			// Add the tx-hash in the rows
			copy(dataForCurrEcdsa[4][:], txHashHi)
			copy(dataForCurrEcdsa[5][:], txHashLo)

			v.FillBytes(buff[:])
			dataForCurrEcdsa[6] = common.Bytes16ToLimbsLe(buff[:16])
			dataForCurrEcdsa[7] = common.Bytes16ToLimbsLe(buff[16:])

			r.FillBytes(buff[:])
			dataForCurrEcdsa[8] = common.Bytes16ToLimbsLe(buff[:16])
			dataForCurrEcdsa[9] = common.Bytes16ToLimbsLe(buff[16:])

			s.FillBytes(buff[:])
			dataForCurrEcdsa[10] = common.Bytes16ToLimbsLe(buff[:16])
			dataForCurrEcdsa[11] = common.Bytes16ToLimbsLe(buff[16:])

			// The success bit // implictly followed by 7 zeroes
			dataForCurrEcdsa[12] = [common.NbLimbU128]field.Element{field.One()}

			// The ecrecover bit (zero)
			dataForCurrEcdsa[13] = [common.NbLimbU128]field.Element{}

			txCounter++
			i += NB_TX_INPUTS

		default:

			// we have run out of inputs.
			break ecdsaLoop
		}

		// Retro-insert the public key in the lower positions from the h, r, s,
		// v that we just parsed.
		pkX, pkY := recoverPk(h, r, s, v)
		pkX.FillBytes(buff[:])
		dataForCurrEcdsa[0] = common.Bytes16ToLimbsLe(buff[:16])
		dataForCurrEcdsa[1] = common.Bytes16ToLimbsLe(buff[16:])

		pkY.FillBytes(buff[:])
		dataForCurrEcdsa[2] = common.Bytes16ToLimbsLe(buff[:16])
		dataForCurrEcdsa[3] = common.Bytes16ToLimbsLe(buff[16:])

		//
		// Properly assigning the data
		//

		// Prepending with zeroes every frame to "skip" the fetching phase
		resIsPublicKey.PushSeqOfZeroes(int(prependZeroCount))
		resGnarkIndex.PushSeqOfZeroes(int(prependZeroCount))
		resGnarkPkIndex.PushSeqOfZeroes(int(prependZeroCount))
		resGnarkData.PushSeqOfZeroes(int(prependZeroCount))

		fmt.Printf("pubX, pubY = %v, %v\n", pkX, pkY)

		fmt.Printf("=========\n")
		printRows(head, dataForCurrEcdsa)
		fmt.Printf("=========\n")

		// Public Key phase
		for i := 0; i < nbRowsPerPublicKey; i++ {
			resIsPublicKey.PushOne()
			resGnarkPkIndex.PushInt(i)
			resGnarkIndex.PushInt(i)
			resGnarkData.PushRow(dataForCurrEcdsa[i][:])
		}

		// Other phase
		for i := nbRowsPerPublicKey; i < nbRowsPerGnarkPushing; i++ {
			resIsPublicKey.PushZero()
			resGnarkPkIndex.PushZero()
			resGnarkIndex.PushInt(i)
			resGnarkData.PushRow(dataForCurrEcdsa[i][:])
		}
	}

	resIsPublicKey.PadAndAssign(run)
	resGnarkIndex.PadAndAssign(run)
	resGnarkPkIndex.PadAndAssign(run)
	resGnarkData.PadAssignZero(run)

}

func (d *UnalignedGnarkData) assignHelperColumns(run *wizard.ProverRuntime, src *unalignedGnarkDataSource) {
	sourceSource := run.GetColumn(src.Source.GetColID())
	sourcIsFetching := run.GetColumn(src.IsFetching.GetColID())
	sourceIsPushing := run.GetColumn(src.IsPushing.GetColID())
	sourceIsData := run.GetColumn(src.IsData.GetColID())
	sourceIsPublicKey := run.GetColumn(d.IsPublicKey.GetColID())
	sourceGnarkIndex := run.GetColumn(d.GnarkIndex.GetColID())
	if sourceSource.Len() != d.Size || sourcIsFetching.Len() != d.Size || sourceIsPushing.Len() != d.Size || sourceIsPublicKey.Len() != d.Size {
		utils.Panic("unexpected source length")
	}
	var resEF, resPP []field.Element
	fe12 := field.NewElement(12)
	for i := 0; i < d.Size; i++ {
		source := sourceSource.Get(i)
		isFetching := sourcIsFetching.Get(i)
		isPushing := sourceIsPushing.Get(i)
		isData := sourceIsData.Get(i)
		isPublicKey := sourceIsPublicKey.Get(i)
		gnarkIndex := sourceGnarkIndex.Get(i)
		if source.Cmp(&SOURCE_ECRECOVER) == 0 && isFetching.IsOne() && isData.IsOne() {
			resEF = append(resEF, field.NewElement(1))
		} else {
			resEF = append(resEF, field.NewElement(0))
		}
		if source.Cmp(&SOURCE_ECRECOVER) == 0 && isPushing.IsOne() && isPublicKey.IsZero() && gnarkIndex.Cmp(&fe12) < 0 {
			resPP = append(resPP, field.NewElement(1))
		} else {
			resPP = append(resPP, field.NewElement(0))
		}
	}
	run.AssignColumn(d.IsEcrecoverAndFetching.GetColID(), smartvectors.RightZeroPadded(resEF, d.Size))
	run.AssignColumn(d.IsNotPublicKeyAndPushing.GetColID(), smartvectors.RightZeroPadded(resPP, d.Size))
}

func (d *UnalignedGnarkData) csDataIds(comp *wizard.CompiledIOP) {
	d.IsIndex0, d.IsIndex0Act = dedicated.IsZero(comp, d.GnarkIndex).GetColumnAndProverAction()
	d.IsIndex4, d.IsIndex4Act = dedicated.IsZero(comp, sym.Sub(d.GnarkIndex, 4)).GetColumnAndProverAction()
	d.IsIndex5, d.IsIndex5Act = dedicated.IsZero(comp, sym.Sub(d.GnarkIndex, 5)).GetColumnAndProverAction()
	d.IsIndex13, d.IsIndex13Act = dedicated.IsZero(comp, sym.Sub(d.GnarkIndex, 13)).GetColumnAndProverAction()
}

func (d *UnalignedGnarkData) csIndex(comp *wizard.CompiledIOP, src *unalignedGnarkDataSource) {
	// that the index is properly increasing. Additionally, we have the
	// gnarkIndexIs13 bit which allows to restart the counting.

	// if IS_FETCHING, then is always 0
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "INDEX_FETCHING_0"),
		sym.Mul(src.IsFetching, d.GnarkIndex),
	)
	// if IS_PUSHING and previous was IS_FETCHING, then is 0
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "INDEX_PUSHING_START"),
		sym.Mul(
			src.IsPushing,
			column.Shift(src.IsFetching, -1),
			sym.Sub(1, d.IsIndex0)),
	)
	// if IS_PUSHING and previous was not IS_FETCHING, then it is previous+1
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "INDEX_PUSHING_INC"),
		sym.Mul(
			src.IsPushing,
			sym.Sub(1, column.Shift(src.IsFetching, -1)),
			sym.Sub(1, sym.Sub(d.GnarkIndex, column.Shift(d.GnarkIndex, -1))),
		),
	)

	// PUBLIC_KEY_INDEX = IS_PUBLIC_KEY * GNARK_INDEX
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "PUBLIC_KEY_INDEX"),
		sym.Sub(d.GnarkPublicKeyIndex, sym.Mul(d.IsPublicKey, d.GnarkIndex)),
	)
}

func (d *UnalignedGnarkData) csProjectionEcRecover(comp *wizard.CompiledIOP, src *unalignedGnarkDataSource) {
	// masks are correctly computed
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "MASKS_A"),
		sym.Sub(d.IsEcrecoverAndFetching, sym.Mul(
			src.IsFetching,
			sym.Sub(1, src.Source),
			src.IsData,
		)),
	)
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "MASKS_B"),
		sym.Sub(d.IsNotPublicKeyAndPushing, sym.Mul(
			sym.Sub(1, d.IsPublicKey),
			sym.Sub(1, src.Source),
			src.IsPushing,
			sym.Sub(1, d.IsIndex13),                  // last is ecrecoverbit
			sym.Sub(1, column.Shift(d.IsIndex13, 1)), // one before last is successbit
		)),
	)
	// that we have projected correctly ecrecover
	comp.InsertProjection(
		ifaces.QueryIDf("%v_PROJECT_ECRECOVER", NAME_UNALIGNED_GNARKDATA),
		query.ProjectionInput{
			ColumnA: src.Limb[:],
			ColumnB: d.GnarkData[:],
			FilterA: d.IsEcrecoverAndFetching,
			FilterB: d.IsNotPublicKeyAndPushing,
		},
	)
}

func (d *UnalignedGnarkData) csTxHash(comp *wizard.CompiledIOP, src *unalignedGnarkDataSource) {
	// that we have projected correctly txHashHi and txHashLo
	for i := 0; i < common.NbLimbU128; i++ {
		comp.InsertGlobal(
			ROUND_NR,
			ifaces.QueryIDf("%v_%v_%d", NAME_UNALIGNED_GNARKDATA, "TXHASH_HI", i),
			sym.Mul(d.IsIndex4, src.Source, sym.Sub(d.GnarkData[i], src.TxHash[i])),
		)

		comp.InsertGlobal(
			ROUND_NR,
			ifaces.QueryIDf("%v_%v_%d", NAME_UNALIGNED_GNARKDATA, "TXHASH_LO", i),
			sym.Mul(d.IsIndex5, src.Source, sym.Sub(d.GnarkData[i], src.TxHash[i+common.NbLimbU128])),
		)
	}
}

func (d *UnalignedGnarkData) csTxEcRecoverBit(comp *wizard.CompiledIOP, src *unalignedGnarkDataSource) {
	// that we set ecrecoverbit correctly for txhash (==0). We do not care for
	// ecrecover as we would be additionally restricting the inputs otherwise
	// and would not be able to solve for valid inputs if the input is valid.

	// additionally, we do not have to binary constrain as it is already
	// enforced inside gnark circuit

	for i := 0; i < common.NbLimbU128; i++ {
		comp.InsertGlobal(
			ROUND_NR,
			ifaces.QueryIDf("%v_%v_%d", NAME_UNALIGNED_GNARKDATA, "ECRECOVERBIT", i),
			sym.Mul(d.IsIndex13, src.Source, d.GnarkData[i]),
		)
	}
}
