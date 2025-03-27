package ecdsa

import (
	"math/big"

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
)

type UnalignedGnarkData struct {
	IsPublicKey         ifaces.Column
	GnarkIndex          ifaces.Column
	GnarkPublicKeyIndex ifaces.Column
	GnarkData           ifaces.Column

	// auxiliary columns
	isIndex0     ifaces.Column
	isIndex0Act  wizard.ProverAction
	isIndex5     ifaces.Column
	isIndex5Act  wizard.ProverAction
	isIndex4     ifaces.Column
	isIndex4Act  wizard.ProverAction
	isIndex13    ifaces.Column
	isIndex13Act wizard.ProverAction

	isEcrecoverAndFetching   ifaces.Column
	isNotPublicKeyAndPushing ifaces.Column

	size int
}

type unalignedGnarkDataSource struct {
	Source     ifaces.Column
	IsActive   ifaces.Column
	IsPushing  ifaces.Column
	IsFetching ifaces.Column
	IsData     ifaces.Column
	IsRes      ifaces.Column
	Limb       ifaces.Column
	SuccessBit ifaces.Column
	TxHashHi   ifaces.Column
	TxHashLo   ifaces.Column
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
		GnarkData:           createCol("GNARK_DATA"),

		isEcrecoverAndFetching:   createCol("IS_ECRECOVER_AND_FETCHING"),
		isNotPublicKeyAndPushing: createCol("IS_NOT_PUBLIC_KEY_AND_PUSHING"),

		size: size,
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
	d.isIndex0Act.Run(run)
	d.isIndex4Act.Run(run)
	d.isIndex5Act.Run(run)
	d.isIndex13Act.Run(run)
}

func (d *UnalignedGnarkData) assignUnalignedGnarkData(run *wizard.ProverRuntime, src *unalignedGnarkDataSource, txSigs TxSignatureGetter) {

	// copies data from the ecrecover part and txn part. Then it also computes
	// the public key values and stores them in the corresponding rows.
	var (
		sourceSource     = run.GetColumn(src.Source.GetColID())
		sourceIsActive   = run.GetColumn(src.IsActive.GetColID())
		sourceLimb       = run.GetColumn(src.Limb.GetColID())
		sourceSuccessBit = run.GetColumn(src.SuccessBit.GetColID())
		sourceTxHashHi   = run.GetColumn(src.TxHashHi.GetColID())
		sourceTxHashLo   = run.GetColumn(src.TxHashLo.GetColID())
	)

	if sourceSource.Len() != d.size || sourceIsActive.Len() != d.size || sourceLimb.Len() != d.size || sourceSuccessBit.Len() != d.size || sourceTxHashHi.Len() != d.size || sourceTxHashLo.Len() != d.size {
		panic("unexpected source length")
	}

	var resIsPublicKey, resGnarkIndex, resGnarkPkIndex, resGnarkData []field.Element
	txCount := 0

	for i := 0; i < d.size; {

		var (
			isActive         = sourceIsActive.Get(i)
			source           = sourceSource.Get(i)
			rows             = make([]field.Element, nbRowsPerGnarkPushing)
			buf              [32]byte
			prehashedMsg     [32]byte
			r, s, v          = new(big.Int), new(big.Int), new(big.Int)
			err              error
			prependZeroCount uint
		)

		if isActive.IsOne() && source.Cmp(&SOURCE_ECRECOVER) == 0 {
			prependZeroCount = nbRowsPerEcRecFetching
			// we copy the data from ecrecover
			rows[12] = sourceSuccessBit.Get(i)
			rows[13] = field.NewElement(1)

			// copy h0, h1, r0, r1, s0, s1, v0, v1
			for j := 0; j < 8; j++ {
				rows[4+j] = sourceLimb.Get(i + j)
			}
			txHighBts := rows[4].Bytes()
			txLowBts := rows[5].Bytes()
			copy(prehashedMsg[:16], txHighBts[16:])
			copy(prehashedMsg[16:], txLowBts[16:])
			v0Bts := rows[6].Bytes()
			v1Bts := rows[7].Bytes()
			copy(buf[:16], v0Bts[16:])
			copy(buf[16:], v1Bts[16:])
			v.SetBytes(buf[:])
			r0Bts := rows[8].Bytes()
			r1Bts := rows[9].Bytes()
			copy(buf[:16], r0Bts[16:])
			copy(buf[16:], r1Bts[16:])
			r.SetBytes(buf[:])
			s0Bts := rows[10].Bytes()
			s1Bts := rows[11].Bytes()
			copy(buf[:16], s0Bts[16:])
			copy(buf[16:], s1Bts[16:])
			s.SetBytes(buf[:])

			i += NB_ECRECOVER_INPUTS
		} else if isActive.IsOne() && source.Cmp(&SOURCE_TX) == 0 {
			prependZeroCount = nbRowsPerTxSignFetching
			// we copy the data from the transcation
			rows[12] = field.NewElement(1) // always succeeds, we only include valid transactions
			rows[13] = field.NewElement(0)

			// copy txHashHi, txHashLo
			rows[4] = sourceTxHashHi.Get(i)
			rows[5] = sourceTxHashLo.Get(i)

			// get r, s, v corresponding to the transaction hash from the provider
			txLow := sourceTxHashLo.Get(i)
			txHigh := sourceTxHashHi.Get(i)
			txLowBts := txLow.Bytes()
			txHighBts := txHigh.Bytes()
			copy(prehashedMsg[:16], txHighBts[16:])
			copy(prehashedMsg[16:], txLowBts[16:])
			r, s, v, err = txSigs(txCount, prehashedMsg[:])
			if err != nil {
				utils.Panic("error getting tx-signature err=%v, txNum=%v", err, txCount)
			}
			v.FillBytes(buf[:])
			rows[6].SetBytes(buf[:16])
			rows[7].SetBytes(buf[16:])
			r.FillBytes(buf[:])
			rows[8].SetBytes(buf[:16])
			rows[9].SetBytes(buf[16:])
			s.FillBytes(buf[:])
			rows[10].SetBytes(buf[:16])
			rows[11].SetBytes(buf[16:])

			i += NB_TX_INPUTS
			txCount++
		} else {
			// we have run out of inputs.
			break
		}
		// compute the expected public key
		var pk ecdsa.PublicKey
		if !v.IsUint64() {
			utils.Panic("v is not a uint64")
		}
		err = pk.RecoverFrom(prehashedMsg[:], uint(v.Uint64()-27), r, s)
		if err != nil {
			utils.Panic("error recovering public: err=%v v=%v r=%v s=%v", err.Error(), v.Uint64()-27, r.String(), s.String())
		}
		pkx := pk.A.X.Bytes()
		rows[0].SetBytes(pkx[:16])
		rows[1].SetBytes(pkx[16:])
		pky := pk.A.Y.Bytes()
		rows[2].SetBytes(pky[:16])
		rows[3].SetBytes(pky[16:])

		resIsPublicKey = append(resIsPublicKey, make([]field.Element, prependZeroCount)...)
		for i := 0; i < 4; i++ {
			resIsPublicKey = append(resIsPublicKey, field.NewElement(1))
		}
		resIsPublicKey = append(resIsPublicKey, make([]field.Element, nbRowsPerGnarkPushing-4)...)

		resGnarkIndex = append(resGnarkIndex, make([]field.Element, prependZeroCount)...)
		resGnarkPkIndex = append(resGnarkPkIndex, make([]field.Element, prependZeroCount)...)
		for i := 0; i < nbRowsPerPublicKey; i++ {
			resGnarkPkIndex = append(resGnarkPkIndex, field.NewElement(uint64(i)))
			resGnarkIndex = append(resGnarkIndex, field.NewElement(uint64(i)))
		}
		for i := nbRowsPerPublicKey; i < nbRowsPerGnarkPushing; i++ {
			resGnarkPkIndex = append(resGnarkPkIndex, field.NewElement(0))
			resGnarkIndex = append(resGnarkIndex, field.NewElement(uint64(i)))
		}

		resGnarkData = append(resGnarkData, make([]field.Element, prependZeroCount)...)
		resGnarkData = append(resGnarkData, rows...)
	}
	// pad the vectors to the full size. It is expected in the hashing module
	// that the underlying vectors have same length.
	resIsPublicKey = append(resIsPublicKey, make([]field.Element, d.size-len(resIsPublicKey))...)
	resGnarkIndex = append(resGnarkIndex, make([]field.Element, d.size-len(resGnarkIndex))...)
	resGnarkPkIndex = append(resGnarkPkIndex, make([]field.Element, d.size-len(resGnarkPkIndex))...)
	resGnarkData = append(resGnarkData, make([]field.Element, d.size-len(resGnarkData))...)
	run.AssignColumn(d.IsPublicKey.GetColID(), smartvectors.RightZeroPadded(resIsPublicKey, d.size))
	run.AssignColumn(d.GnarkIndex.GetColID(), smartvectors.RightZeroPadded(resGnarkIndex, d.size))
	run.AssignColumn(d.GnarkPublicKeyIndex.GetColID(), smartvectors.RightZeroPadded(resGnarkPkIndex, d.size))
	run.AssignColumn(d.GnarkData.GetColID(), smartvectors.RightZeroPadded(resGnarkData, d.size))
}

func (d *UnalignedGnarkData) assignHelperColumns(run *wizard.ProverRuntime, src *unalignedGnarkDataSource) {
	sourceSource := run.GetColumn(src.Source.GetColID())
	sourcIsFetching := run.GetColumn(src.IsFetching.GetColID())
	sourceIsPushing := run.GetColumn(src.IsPushing.GetColID())
	sourceIsData := run.GetColumn(src.IsData.GetColID())
	sourceIsPublicKey := run.GetColumn(d.IsPublicKey.GetColID())
	sourceGnarkIndex := run.GetColumn(d.GnarkIndex.GetColID())
	if sourceSource.Len() != d.size || sourcIsFetching.Len() != d.size || sourceIsPushing.Len() != d.size || sourceIsPublicKey.Len() != d.size {
		utils.Panic("unexpected source length")
	}
	var resEF, resPP []field.Element
	fe12 := field.NewElement(12)
	for i := 0; i < d.size; i++ {
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
	run.AssignColumn(d.isEcrecoverAndFetching.GetColID(), smartvectors.RightZeroPadded(resEF, d.size))
	run.AssignColumn(d.isNotPublicKeyAndPushing.GetColID(), smartvectors.RightZeroPadded(resPP, d.size))
}

func (d *UnalignedGnarkData) csDataIds(comp *wizard.CompiledIOP) {
	d.isIndex0, d.isIndex0Act = dedicated.IsZero(comp, d.GnarkIndex)
	d.isIndex4, d.isIndex4Act = dedicated.IsZero(comp, sym.Sub(d.GnarkIndex, 4))
	d.isIndex5, d.isIndex5Act = dedicated.IsZero(comp, sym.Sub(d.GnarkIndex, 5))
	d.isIndex13, d.isIndex13Act = dedicated.IsZero(comp, sym.Sub(d.GnarkIndex, 13))
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
			sym.Sub(1, d.isIndex0)),
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
		sym.Sub(d.isEcrecoverAndFetching, sym.Mul(
			src.IsFetching,
			sym.Sub(1, src.Source),
			src.IsData,
		)),
	)
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "MASKS_B"),
		sym.Sub(d.isNotPublicKeyAndPushing, sym.Mul(
			sym.Sub(1, d.IsPublicKey),
			sym.Sub(1, src.Source),
			src.IsPushing,
			sym.Sub(1, d.isIndex13),                  // last is ecrecoverbit
			sym.Sub(1, column.Shift(d.isIndex13, 1)), // one before last is successbit
		)),
	)
	// that we have projected correctly ecrecover
	comp.InsertProjection(
		ifaces.QueryIDf("%v_PROJECT_ECRECOVER", NAME_UNALIGNED_GNARKDATA),
		query.ProjectionInput{ColumnA: []ifaces.Column{src.Limb},
			ColumnB: []ifaces.Column{d.GnarkData},
			FilterA: d.isEcrecoverAndFetching,
			FilterB: d.isNotPublicKeyAndPushing})
}

func (d *UnalignedGnarkData) csTxHash(comp *wizard.CompiledIOP, src *unalignedGnarkDataSource) {
	// that we have projected correctly txHashHi and txHashLo
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "TXHASH_HI"),
		sym.Mul(d.isIndex4, src.Source, sym.Sub(d.GnarkData, src.TxHashHi)),
	)
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "TXHASH_LO"),
		sym.Mul(d.isIndex5, src.Source, sym.Sub(d.GnarkData, src.TxHashLo)),
	)
}

func (d *UnalignedGnarkData) csTxEcRecoverBit(comp *wizard.CompiledIOP, src *unalignedGnarkDataSource) {
	// that we set ecrecoverbit correctly for txhash (==0). We do not care for
	// ecrecover as we would be additionally restricting the inputs otherwise
	// and would not be able to solve for valid inputs if the input is valid.

	// additionally, we do not have to binary constrain as it is already
	// enforced inside gnark circuit
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_%v", NAME_UNALIGNED_GNARKDATA, "ECRECOVERBIT"),
		sym.Mul(d.isIndex13, src.Source, d.GnarkData),
	)
}
