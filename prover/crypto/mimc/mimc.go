package mimc

import (
	"hash"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/frontend"
)

/*
	This package wraps the mimc package of gnark-crypto and implements
	some extra methods.
*/

func NewMiMC() hash.Hash {
	return mimc.NewMiMC()
}

// Constants collects the MiMC constants parsed as field elements
var Constants []field.Element = func() []field.Element {
	bigConsts := mimc.GetConstants()
	res := make([]field.Element, len(bigConsts))
	for i := range res {
		res[i].SetBigInt(&bigConsts[i])
	}
	return res
}()

// Applies the MiMC permutation to a given block
func BlockCompression(oldState, block field.Element) (newState field.Element) {

	res := block
	var tmp field.Element

	// s <- (s + old + c)^5
	for i := range Constants {
		// We don't use the loop value of Constant to explictly
		// show the linter that we are not mutating the loop value.
		c := Constants[i]
		res.Add(&res, &c)
		res.Add(&res, &oldState)
		tmp.Square(&res)
		tmp.Square(&tmp)
		res.Mul(&tmp, &res)
	}

	res.Add(&res, &oldState).Add(&res, &oldState).Add(&res, &block)
	return res
}

// Applies the MiMC permutation to a given block
func GnarkBlockCompression(api frontend.API, oldState, block frontend.Variable) (newState frontend.Variable) {

	res := block
	var tmp frontend.Variable

	// s <- (s + old + c)^5
	for _, c := range Constants {
		res = api.Add(res, c)
		res = api.Add(res, oldState)
		tmp = api.Mul(res, res)
		tmp = api.Mul(tmp, tmp)
		res = api.Mul(tmp, res)
	}

	res = api.Add(res, oldState)
	res = api.Add(res, oldState)
	res = api.Add(res, block)

	return res
}

// Hash a vector of field elements
func HashVec(v []field.Element) (h field.Element) {
	state := NewMiMC()
	for i := range v {
		vBytes := v[i].Bytes()
		state.Write(vBytes[:])
	}
	h.SetBytes(state.Sum(nil))
	return
}
