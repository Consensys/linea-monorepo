package verifiercol

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// compile check to enforce the struct to belong to the corresponding interface
var _ VerifierCol = FromIntVecCoin{}

// Represents a columns instantiated by the values of a
// a random indices list.
type FromIntVecCoin struct {
	info    coin.Info
	padding struct {
		isPadded   bool
		paddingVal field.Element
		fullLen    int
	}
	// The splitting is always applied AFTER the padding
	splitting struct {
		isSplit  bool
		from, to int
	}
}

// Options to pass to FIVC
type FivcOp func(*FromIntVecCoin)

// Construct a new column from a `IntegerVec` coin
func NewFromIntVecCoin(comp *wizard.CompiledIOP, info coin.Info, ops ...FivcOp) ifaces.Column {

	// Sanity-checks the coin to have the right type
	if info.Type != coin.IntegerVec {
		utils.Panic("FromIntVecCoin : expected type integer vec")
	}

	// Initialize a naked fivc
	res := FromIntVecCoin{info: info}

	// And apply the options
	for _, op := range ops {
		op(&res)
	}

	// If the size is not a power of two. We will have
	// issues when computing polynomial evaluations in
	// Lagrange basis. So we prefer to forbid it. However
	// the check should be run after applying the options.
	if !utils.IsPowerOfTwo(res.Size()) {
		utils.Panic("The size should be a power of two. It was %v", res.Size())
	}

	return res
}

// Passes a padding value to the Fivc
func RightPadZeroToNextPowerOfTwo(fivc *FromIntVecCoin) {
	// For sanity, the paddding can never happen over a split FIVC
	if fivc.padding.isPadded {
		utils.Panic("tried to pad a split FIVC vector : %v", fivc.info.Name)
	}

	// Skip it if the length is already a power of two
	if utils.IsPowerOfTwo(fivc.info.Size) {
		return
	}

	fivc.padding.isPadded = true
	fivc.padding.fullLen = utils.NextPowerOfTwo(fivc.info.Size)
}

// Returns the round of definition of the column
func (fivc FromIntVecCoin) Round() int {
	return fivc.info.Round
}

// Returns a generic name from the column. Defined from the coin's.
func (fivc FromIntVecCoin) GetColID() ifaces.ColID {

	switch {
	case fivc.splitting.isSplit:
		//
		return ifaces.ColIDf(
			"FIVC_%v_SPLIT_%v_%v",
			fivc.info.Name,
			fivc.splitting.from,
			fivc.splitting.to,
		)
	case fivc.padding.isPadded:
		//
		return ifaces.ColIDf(
			"FIVC_%v_PADDED_%v_%v",
			fivc.info.Name,
			fivc.padding.paddingVal.String(),
			fivc.padding.fullLen,
		)
	default:
		return ifaces.ColIDf("FIVC_%v", fivc.info.Name)
	}
}

// Always return true. We sanity-check the existence of the
// random coin prior to constructing the object.
func (fivc FromIntVecCoin) MustExists() {}

// Returns the size of the integer vec coin
func (fivc FromIntVecCoin) Size() int {

	// Sanity-checks the coin to have the right type
	if fivc.info.Type != coin.IntegerVec {
		utils.Panic("FromIntVecCoin : expected type integer vec")
	}

	// If it is a split FIVC, then the priority is given to the
	// splitting parameters to deduce the size
	if fivc.splitting.isSplit {
		return fivc.splitting.to - fivc.splitting.from
	}

	// Else, this goes to the padding argment if set
	if fivc.padding.isPadded {
		return fivc.padding.fullLen
	}

	// Else, it is in the specification of the coin
	return fivc.info.Size
}

// Returns the coin's value as a column assignment
func (fivc FromIntVecCoin) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {

	ints := run.GetRandomCoinIntegerVec(fivc.info.Name)

	// Sanity-check
	if len(ints) != fivc.info.Size {
		utils.Panic("the size does not match : %v != %v", len(ints), fivc.info.Size)
	}

	// Cast as field elements
	fields := make([]field.Element, fivc.info.Size)
	for i := range fields {
		fields[i].SetUint64(uint64(ints[i]))
	}

	// If the padding is enabled, then apply it
	var res smartvectors.SmartVector = smartvectors.NewRegular(fields)
	if fivc.padding.isPadded {
		res = smartvectors.RightPadded(fields, fivc.padding.paddingVal, fivc.padding.fullLen)
	}

	// If the splitting is enabled, then apply it
	if fivc.splitting.isSplit {
		res = res.SubVector(fivc.splitting.from, fivc.splitting.to)
	}

	return res
}

// Returns the coin's value as a column assignment
func (fivc FromIntVecCoin) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {

	res := run.GetRandomCoinIntegerVec(fivc.info.Name)

	// Sanity-check
	if len(res) != fivc.info.Size {
		utils.Panic("the size does not match : %v != %v", len(res), fivc.info.Size)
	}

	// If the padding is enabled, then apply it
	if padding := fivc.padding; padding.isPadded {
		padded := make([]frontend.Variable, 0, padding.fullLen)
		padded = append(padded, res...)
		for i := len(res); i < padding.fullLen; i++ {
			padded = append(padded, padding.paddingVal)
		}
		res = padded
	}

	// if the splitting is required, then apply it
	if splitting := fivc.splitting; splitting.isSplit {
		res = res[splitting.from:splitting.to]
	}

	return res
}

// Returns a particular position of the coin value
func (fivc FromIntVecCoin) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return fivc.GetColAssignment(run).Get(pos)
}

// Returns a particular position of the coin value
func (fivc FromIntVecCoin) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	return fivc.GetColAssignmentGnark(run)[pos]
}

// Since the column is directly defined from the
// values of a random coin it does not count as a
// composite column.
func (fivc FromIntVecCoin) IsComposite() bool {
	return false
}

// Returns the name of the column.
func (fivc FromIntVecCoin) String() string {
	return string(fivc.GetColID())
}

// Splits the column and return a handle of it
func (fivc FromIntVecCoin) Split(comp *wizard.CompiledIOP, from, to int) ifaces.Column {

	if to < from || to-from > fivc.Size() {
		utils.Panic("Can't split %++v into [%v, %v]", fivc, from, to)
	}

	// Account for the fact that the underlying fivc may be already
	// split. In that case, we need offset the from and the to.
	if splitting := fivc.splitting; splitting.isSplit {
		from, to = splitting.from+from, splitting.from+to
	}

	// Copy the underlying fivc, and assigns the new from and to
	res := fivc
	// if not marked as split, then from and to will simply be ignored
	res.splitting.isSplit = true
	res.splitting.from = from
	res.splitting.to = to

	// sanity-check : the outgoing slice should have the expected size
	if res.Size() != to-from {
		utils.Panic("post-condition failed : res.Size()=%v but from=%v and to=%v", res.Size(), from, to)
	}

	return res
}
