package testing

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// AlterProof randomly mutates a proof object so that it is no longer valid.
// This is used to turn a valid proof into an invalid one. The function
// performs only one mutation. The mutations are deterministic in the sense
// that it relies in an internal PRNG.
func AlterProof(proof *wizard.Proof) {

	methods := []func(proof *wizard.Proof){
		func(proof *wizard.Proof) {

			var (
				qNames   = proof.QueriesParams.ListAllKeys()
				qName    = qNames[rng.Intn(len(qNames))]
				query    = proof.QueriesParams.MustGet(qName)
				newQuery = randomlyMutateQueryParams(query)
			)

			proof.QueriesParams.Update(qName, newQuery)

		},
		func(proof *wizard.Proof) {

			var (
				cNames = proof.Messages.ListAllKeys()
				cName  = cNames[rng.Intn(len(cNames))]
				col    = proof.Messages.MustGet(cName).IntoRegVecSaveAlloc()
				newCol = randomlyMutateVector(col)
			)

			proof.Messages.Update(cName, smartvectors.NewRegular(newCol))
		},
	}

	method := methods[rng.Intn(len(methods))]
	method(proof)
}

func randomlyMutateQueryParams(params ifaces.QueryParams) ifaces.QueryParams {

	switch params := params.(type) {

	case query.UnivariateEvalParams:

		methods := []func(params query.UnivariateEvalParams) ifaces.QueryParams{
			func(params query.UnivariateEvalParams) ifaces.QueryParams {
				randomlyMutateField(&params.X)
				return params
			},
			func(params query.UnivariateEvalParams) ifaces.QueryParams {
				params.Ys = randomlyMutateVector(params.Ys)
				return params
			},
		}

		method := methods[rng.Intn(len(methods))]
		return method(params)

	case query.InnerProductParams:

		params.Ys = randomlyMutateVector(params.Ys)
		return params

	case query.HornerParams:

		methods := []func(params query.HornerParams) ifaces.QueryParams{
			func(params query.HornerParams) ifaces.QueryParams {
				randomlyMutateField(&params.FinalResult)
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				params.Parts[pos].N0++
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				params.Parts[pos].N0--
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				if params.Parts[pos].N0 > 0 {
					params.Parts[pos].N0 = 0
				} else {
					params.Parts[pos].N0++
				}
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				if params.Parts[pos].N0 < params.Parts[pos].N1 {
					params.Parts[pos].N0 = params.Parts[pos].N1
				} else {
					params.Parts[pos].N0++
				}
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				params.Parts[pos].N1++
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				params.Parts[pos].N1--
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				if params.Parts[pos].N1 > 0 {
					params.Parts[pos].N1 = 0
				} else {
					params.Parts[pos].N1++
				}
				return params
			},
			func(params query.HornerParams) ifaces.QueryParams {
				pos := rng.Intn(len(params.Parts))
				if params.Parts[pos].N1 < params.Parts[pos].N0 {
					params.Parts[pos].N1 = params.Parts[pos].N0
				} else {
					params.Parts[pos].N1++
				}
				return params
			},
		}

		method := methods[rng.Intn(len(methods))]
		return method(params)

	case query.LogDerivSumParams:

		randomlyMutateField(&params.Sum)
		return params

	case query.GrandProductParams:

		randomlyMutateField(&params.Y)
		return params

	case query.LocalOpeningParams:

		randomlyMutateField(&params.Y)
		return params
	}

	panic("unreachable")

}

func randomlyMutateField(f *field.Element) {

	methods := []func(f *field.Element){
		func(f *field.Element) { f.SetZero() },
		func(f *field.Element) { f.SetOne() },
		func(f *field.Element) { (*f) = field.PseudoRand(rng) },
		func(f *field.Element) { f.Neg(f) },
		func(f *field.Element) { f.Add(f, new(field.Element).SetOne()) },
	}

	oldVal := *f

	for oldVal == *f {
		method := methods[rng.Intn(len(methods))]
		method(f)
	}
}

func randomlyMutateVector(v []field.Element) []field.Element {

	methods := []func(v []field.Element) []field.Element{
		func(v []field.Element) []field.Element {
			return v[:len(v)/2]
		},
		func(v []field.Element) []field.Element {
			return v[len(v)/2:]
		},
		func(v []field.Element) []field.Element {
			return vector.PseudoRand(rng, len(v))
		},
		func(v []field.Element) []field.Element {
			return make([]field.Element, len(v))
		},
		func(v []field.Element) []field.Element {
			res := []field.Element{}
			res = append(res, v...)
			res = append(res, v...)
			return res
		},
		func(v []field.Element) []field.Element {
			res := make([]field.Element, len(v))
			vector.ScalarMul(res, v, field.NewElement(2))
			return res
		},
		func(v []field.Element) []field.Element {
			res := make([]field.Element, len(v))
			vector.ScalarMul(res, v, field.PseudoRand(rng))
			return res
		},
		func(v []field.Element) []field.Element {
			return field.BatchInvert(v)
		},
		func(v []field.Element) []field.Element {
			res := append([]field.Element{}, v...)
			randomlyMutateField(&res[rng.Intn(len(res))])
			return res
		},
	}

	newVal := v
	for vector.Equal(newVal, v) {
		method := methods[rng.Intn(len(methods))]
		newVal = method(v)
	}

	return newVal
}
