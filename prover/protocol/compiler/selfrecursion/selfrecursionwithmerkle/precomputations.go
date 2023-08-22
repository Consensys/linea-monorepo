package selfrecursionwithmerkle

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Registers the polynomial I(X) = (0, 1, 2, 3, 4, 5, ..., nBcol-1)
// and the SIS key chunks.
//
// Say we have a random ring-SIS key with dimension
//
//	(A0(X),A1(X),A2(X),…,An−1(X))
//
// and that I need k polynomials to hash a field element. We need
// k≥1, so that sets some conditions on the ring-SIS instance. If
// we plan to commit to m polynomials at once at some point of the
// protocol and that at that point I would already have committed
// to m′ polynomials. In that case, the key shard Ah would be
//
//	(0(X)×km′) ∥ (A0(X),…,Akm(X)) ∥ (0(X)…up to kmf)
//
// Namely, all key shards consists of the first entries of A with
// some offset and are zeroes everywhere else.
func (ctx *SelfRecursionCtx) Precomputations() {
	ctx.registersI()
	ctx.registersAh()
}

// Registers the polynomial I(X)
func (ctx *SelfRecursionCtx) registersI() {
	nBColEncoded := ctx.VortexCtx.NumEncodedCols()
	i := make([]field.Element, nBColEncoded)
	for k := range i {
		i[k].SetUint64(uint64(k))
	}
	ctx.Columns.I = ctx.comp.InsertPrecomputed(ctx.iName(nBColEncoded), smartvectors.NewRegular(i))
}

// Registers the key shards, since some rounds are dried, some of the
// of the entries are nil
func (ctx *SelfRecursionCtx) registersAh() {

	ah := make([]ifaces.Column, ctx.VortexCtx.CommitmentsByRounds.Len())

	// Tracks the number of rows already committed as we arrive in this
	// round
	maxSize := utils.NextPowerOfTwo(ctx.VortexCtx.CommittedRowsCount)
	roundStartAt := 0
	for i, comsInRoundsI := range ctx.VortexCtx.CommitmentsByRounds.Inner() {

		// Sanity-check : if coms in rounds has length zero then the
		// associate Dh should be nil
		if (len(comsInRoundsI) == 0) != (ctx.Columns.Rooth[i] == nil) {
			panic("nilness mismatch")
		}

		// Check if there is no rows to commit
		if len(comsInRoundsI) == 0 {
			// and ah[i] is nil
			continue
		}

		// Registers the commitment key (if this matches an existing key
		// then the preexisting precomputed key is reused.
		ah[i] = ctx.comp.InsertPrecomputed(
			ctx.ahName(ctx.SisKey(), roundStartAt, len(comsInRoundsI), maxSize),
			LaidOutKeyChunk(ctx.SisKey(), roundStartAt, len(comsInRoundsI), maxSize),
		)

		// And update the value of the start
		roundStartAt += len(comsInRoundsI)
	}

	ctx.Columns.Ah = ah

	// It's cleaner for us to have len(dH) == len(aH), so we enforces that
	// by `nil` appending the two slices so that they have the same size at
	// the end
	for len(ctx.Columns.Ah) < len(ctx.Columns.Rooth) {
		ctx.Columns.Ah = append(ctx.Columns.Ah, nil)
	}

	for len(ctx.Columns.Rooth) < len(ctx.Columns.Ah) {
		ctx.Columns.Rooth = append(ctx.Columns.Rooth, nil)
	}

	// And normally, they have the same length now
	if len(ctx.Columns.Ah) != len(ctx.Columns.Rooth) {
		panic("ah and dh should have had the same length now")
	}
}

// Returns the laid out keys
func LaidOutKeyChunk(key *ringsis.Key, start, length, maxSize int) smartvectors.SmartVector {

	// Sanity-check : the chunkNo can't be off-bound
	if maxSize < start+length {
		utils.Panic("inconsistent arguments : %v + %v > %v", start, length, maxSize)
	}

	// Case for the last chunk
	laidOutKey := key.LaidOutKey()
	res := make([]field.Element, maxSize*key.NumLimbs())

	// Scales the start and the length according to how many limbs
	// are necessary to encode a field element.
	startAt := start * key.NumLimbs()
	numToWrite := length * key.NumLimbs()

	copy(res[startAt:], laidOutKey[:numToWrite])
	return smartvectors.NewRegular(res)
}
