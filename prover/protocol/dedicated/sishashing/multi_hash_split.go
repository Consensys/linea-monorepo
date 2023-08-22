package sishashing

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/expr_handle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/functionals"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizardutils"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

// Verify a batch of sis hash for split columns.
//
// Say that we have 30 columns to hash and theses columns are all split
// in chunks 4 all starting at positions [0, 15, 80, 174] (the position)
// For each of these chunk we have a concatenated digest.
//
// The wizard that this function build ensures that these concatenated
// hashes are correctly built.
func MultiSplitSplitCheck(
	// string identifier for the declared instance of RingSIS check
	instanceName string,
	// compiled wizard iop
	comp *wizard.CompiledIOP,
	key *ringsis.Key,
	limbSplitPreimage []ifaces.Column,
	// concatenated hashes
	concatenatedSisHash []ifaces.Column,
	// list of indices giving the position of all indices
	splitIndices []int,
) {

	// Perform a sequence of sanity-checks

	if len(concatenatedSisHash) != len(splitIndices) {
		utils.Panic("there should be as many hashes as there are indices")
	}

	/*
		here the preimage (in limbs) have their size inflated by `key.NumLimbs`
		compared to the size of the preimage in "field" form. Hence, the division
		by key.NumLimbs to obtain the size of the hashes in field element rather
		than in limbs.
	*/
	size := wizardutils.AssertAllHandleSameLength(limbSplitPreimage...) / key.NumLimbs()
	assertCorrectSplitList(splitIndices, size)

	if key.Degree > key.NumLimbs() {
		utils.Panic("UNSUPPORTED, the number of limbs per field is smaller" +
			"than the degree of the sis instance. That means, that we can not" +
			"always `cleanly` cut the sis key")
	}

	/*
		1. Precompute to the laid keys chunk which are all
		of the form:

			[0; N] || A[:M] || [0;N]

		and all have the same size as the .
	*/
	keyChunks := make([]ifaces.Column, len(splitIndices))
	for chunkNo := range splitIndices {
		keyChunks[chunkNo] = comp.InsertPrecomputed(
			PrecomputedKeyChunkName(key, splitIndices, chunkNo, size),
			LaidOutKeyChunk(key, splitIndices, chunkNo, size),
		)
	}

	maxRound := wizardutils.MaxRound(append(limbSplitPreimage, concatenatedSisHash...)...)

	// Range checks the preimage
	for _, preimage := range limbSplitPreimage {
		comp.InsertRange(
			maxRound,
			ifaces.QueryIDf("%v_%v_SHORTNESS", instanceName, preimage.GetColID()),
			preimage,
			1<<key.LogTwoBound,
		)
	}

	/*
		2. When the prover has claimed all the hashes and their
		preimages

			* Declare a coin `mergeKeyCoin`
			* Compute the LC of the alleged digests into a global one
			* Compute the LC of the alleged keys into a global one
			* Declare the dual sis hash of the merged sis hash
	*/
	mergerCoin := comp.InsertCoin(
		maxRound+1,
		coin.Namef("%v_MERGER", instanceName),
		coin.Field,
	)

	mergedSisKey := expr_handle.RandLinCombCol(comp,
		accessors.AccessorFromCoin(mergerCoin),
		keyChunks,
		fmt.Sprintf("%v_MERGED_KEY", instanceName),
	)

	mergedConcatenatedSisHash := expr_handle.RandLinCombCol(comp,
		accessors.AccessorFromCoin(mergerCoin),
		concatenatedSisHash,
		fmt.Sprintf("%v_MERGED_HASHES", instanceName),
	)

	// Introduces a messages correspondind to the sis hash, but
	// modulo X^n - 1. It's not really a hash per se (because this
	// is not binding), but it is used as an advice value.
	mergedConcatenatedDualSisHash := comp.InsertProof(
		maxRound+1,
		ifaces.ColID(fmt.Sprintf("%v_DUAL", instanceName)),
		mergedConcatenatedSisHash.Size(),
	)

	// Assign the dual sis hash, by concatenating the dual sis hash of all columns
	comp.SubProvers.AppendToInner(maxRound+1, func(assi *wizard.ProverRuntime) {

		// Collect the preimages of the witnesses
		limbSplitPreimageWit := make([]smartvectors.SmartVector, len(limbSplitPreimage))
		for i := range limbSplitPreimageWit {
			limbSplitPreimageWit[i] = limbSplitPreimage[i].GetColAssignment(assi)
		}

		// Returns the list of all the hashes modulo X^n - 1
		hashes := hashModXMinusOneSplit(key, limbSplitPreimageWit, splitIndices)
		mergerCoin := assi.GetRandomCoinField(mergerCoin.Name)
		mergedDualHash := smartvectors.PolyEval(hashes, mergerCoin)

		assi.AssignColumn(mergedConcatenatedDualSisHash.GetColID(), mergedDualHash)
	})

	/*
		3. At the same time

			* Declare a coin `collapseCoin`
			* Compute the LC of all the limb preimages into a global one

			* Declare a folder coin
			* Fold the collapsed preimage
			* Fold the mergedKey
			* Declare the inner-product between the two
			* Assigns the result of the inner-product
	*/

	// Declare the folding coin
	folderCoin := comp.InsertCoin(mergedSisKey.Round()+1, coin.Namef("%v_FOLDER", instanceName), coin.Field)
	folder := accessors.AccessorFromCoin(folderCoin)

	// Collapse the preimages into a linear combination
	collapseCoin := comp.InsertCoin(mergedSisKey.Round()+1, coin.Namef("%v_COLLAPSER", instanceName), coin.Field)
	linCombP := expr_handle.RandLinCombCol(
		comp,
		accessors.AccessorFromCoin(collapseCoin),
		limbSplitPreimage)

	// fold the key and the (collapsed preimage)
	foldedKey := functionals.Fold(comp, mergedSisKey, folder, key.Degree)
	foldedPreimage := functionals.Fold(comp, linCombP, folder, key.Degree)

	/*
		4. Declare an inner-product query between the folded
			key and the folded preimage.
	*/

	ipRound := utils.Max(foldedKey.Round(), foldedPreimage.Round())
	ipName := ifaces.QueryIDf("%v_IP", instanceName)
	comp.InsertInnerProduct(ipRound, ipName, foldedKey, []ifaces.Column{foldedPreimage})

	comp.SubProvers.AppendToInner(ipRound, func(assi *wizard.ProverRuntime) {
		// compute the inner-product
		foldedKey := foldedKey.GetColAssignment(assi)           // overshadows the handle
		foldedPreimage := foldedPreimage.GetColAssignment(assi) // overshadows the handle

		y := smartvectors.InnerProduct(foldedKey, foldedPreimage)
		assi.AssignInnerProduct(ipName, y)
	})

	/*
		4. The verifier
			* evaluate the dual sis hash on the folder coin (X) and
				the collapse coin (Y)
			* evaluate the merged sis hash on the folder coin (X) and
				the collapse coin (Y)
			* fetch the result of the inner product and use the following
				consistency check
	*/

	// check the folding of the polynomial is correct
	comp.InsertVerifier(ipRound, func(a *wizard.VerifierRuntime) error {
		mergedConcatenatedSisHashWith := a.GetColumn(mergedConcatenatedSisHash.GetColID())
		mergedConcatenantedDualSisHashWith := a.GetColumn(mergedConcatenatedDualSisHash.GetColID())

		x := folder.GetVal(a)
		t := a.GetRandomCoinField(collapseCoin.Name)
		yAlleged := a.GetInnerProductParams(ipName).Ys[0]
		yDual := smartvectors.EvalCoeffBivariate(mergedConcatenantedDualSisHashWith, x, key.Degree, t)
		yActual := smartvectors.EvalCoeffBivariate(mergedConcatenatedSisHashWith, x, key.Degree, t)

		/*
			If P(X) is of degree 2d

			And
				- Q(X) = P(X) mod X^n - 1
				- R(X) = P(X) mod X^n + 1

			Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
			Here, we can identify at the point x

			yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
		*/
		var xN, xNminus1, xNplus1 field.Element
		one := field.One()
		xN.Exp(x, big.NewInt(int64(key.Degree)))
		xNminus1.Sub(&xN, &one)
		xNplus1.Add(&xN, &one)

		var left, left0, left1, right field.Element
		left0.Mul(&xNplus1, &yDual)
		left1.Mul(&xNminus1, &yActual)
		left.Sub(&left0, &left1)

		right.Double(&yAlleged)

		if left != right {
			return fmt.Errorf("failed the consistency check of the ring-SIS : %v != %v", left.String(), right.String())
		}

		return nil
	}, func(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
		mergedConcatenatedSisHashWit := wvc.GetColumn(mergedConcatenatedSisHash.GetColID())
		mergedConcatenantedDualSisHashWith := wvc.GetColumn(mergedConcatenatedDualSisHash.GetColID())

		x := folder.GetFrontendVariable(api, wvc)
		t := wvc.GetRandomCoinField(collapseCoin.Name)
		yAlleged := wvc.GetInnerProductParams(ipName).Ys[0]

		yDual := poly.GnarkEvalCoeffBivariate(api, mergedConcatenantedDualSisHashWith, x, key.Degree, t)
		yActual := poly.GnarkEvalCoeffBivariate(api, mergedConcatenatedSisHashWit, x, key.Degree, t)

		/*
			If P(X) is of degree 2d

			And
				- Q(X) = P(X) mod X^n - 1
				- R(X) = P(X) mod X^n + 1

			Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
			Here, we can identify at the point x

			yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
		*/
		one := field.One()
		xN := gnarkutil.Exp(api, x, key.Degree)
		xNminus1 := api.Sub(xN, one)
		xNplus1 := api.Add(xN, one)
		left0 := api.Mul(xNplus1, yDual)
		left1 := api.Mul(xNminus1, yActual)
		left := api.Sub(left0, left1)
		right := api.Mul(yAlleged, 2)
		api.AssertIsEqual(left, right)
	})

}

func assertCorrectSplitList(list []int, maxSize int) {
	if list[0] != 0 {
		utils.Panic("the first indice should be zero. Was %v", list[0])
	}

	for i := range list {
		// skip i=0 otherwise i-1 is off-bound
		if i == 0 {
			continue
		}
		// the indices should be increasing
		if list[i] <= list[i-1] {
			utils.Panic("the list should be increasing but list[%v] (= %v) <= list[%v] (= %v)", i, list[i], i-1, list[i-1])
		}
	}

	// it cannot go off-bound of a certain limit
	if list[len(list)-1] >= maxSize {
		utils.Panic("the chunking goes beyond the size of the concatenated column : %v > %v", list[len(list)-1], maxSize)
	}
}

func PrecomputedKeyChunkName(key *ringsis.Key, splitIndices []int, chunkNo, maxSize int) ifaces.ColID {

	// sanity-check : the chunkNo can't be off-bound
	if chunkNo >= len(splitIndices) {
		utils.Panic("chunkNo is off-bound (%v, bound is %v)", chunkNo, len(splitIndices))
	}

	// case for the last chunk
	length := maxSize - splitIndices[chunkNo]
	// case for the non-last chunks
	if chunkNo < len(splitIndices)-1 {
		length = splitIndices[chunkNo+1] - splitIndices[chunkNo]
	}

	subName := SisKeyName(key)
	return ifaces.ColIDf("%v_%v_%v", subName, splitIndices[chunkNo], length)
}

func LaidOutKeyChunk(key *ringsis.Key, splitIndices []int, chunkNo, maxSize int) smartvectors.SmartVector {

	// Sanity-check : the chunkNo can't be off-bound
	if chunkNo >= len(splitIndices) {
		utils.Panic("chunkNo is off-bound (%v, bound is %v)", chunkNo, len(splitIndices))
	}

	// Case for the last chunk
	length := maxSize - splitIndices[chunkNo]
	// case for the non-last chunks
	if chunkNo < len(splitIndices)-1 {
		length = splitIndices[chunkNo+1] - splitIndices[chunkNo]
	}

	// Number of element of the sis key to keep
	laidOutKey := key.LaidOutKey()
	res := make([]field.Element, maxSize*key.NumLimbs())

	startAt := splitIndices[chunkNo] * key.NumLimbs()
	numToWrite := length * key.NumLimbs()

	copy(res[startAt:], laidOutKey[:numToWrite])
	return smartvectors.NewRegular(res)
}

// Obtain the dual sis hash
func hashModXMinusOneSplit(key *ringsis.Key, limbSplitPreimages []smartvectors.SmartVector, splitIndices []int) []smartvectors.SmartVector {

	allHashes := make([]smartvectors.SmartVector, len(splitIndices))

	// Compute all the "sub-hashes" one by one
	for chunkNo := range allHashes {
		resChunk := make([]field.Element, key.Degree*len(limbSplitPreimages))
		for i := range limbSplitPreimages {

			// Case for the last chunk
			length := limbSplitPreimages[0].Len()/key.NumLimbs() - splitIndices[chunkNo]
			// case for the non-last chunks
			if chunkNo < len(splitIndices)-1 {
				length = splitIndices[chunkNo+1] - splitIndices[chunkNo]
			}

			startAt := splitIndices[chunkNo] * key.NumLimbs()
			numToWrite := length * key.NumLimbs()

			chunked := limbSplitPreimages[i].SubVector(startAt, startAt+numToWrite)
			chunkedSlice := smartvectors.IntoRegVec(chunked)

			// Perform the hash and accumulate it in the result
			hashI := key.HashModXnMinus1(chunkedSlice)
			copy(resChunk[i*key.Degree:(i+1)*key.Degree], hashI)
		}

		allHashes[chunkNo] = smartvectors.NewRegular(resChunk)
	}

	return allHashes
}
