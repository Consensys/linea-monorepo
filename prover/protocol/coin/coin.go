package coin

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Wrapper type for naming the coins
type Name string

// Utility function to format names
func Namef(s string, args ...interface{}) Name {
	return Name(fmt.Sprintf(s, args...))
}

// Metadata around the random coin
type Info struct {
	Type
	// Set if applicable (for instance, IntegerVec)
	Size int
	// Upper-bound (if applicable, for instance for integers)
	UpperBound int
	// Name of the coin
	Name Name
	// Round at which the coin was declared
	Round int
}

// Type of random coin
type Type int

const (
	Field Type = iota
	IntegerVec
)

/*
Sample a random coin, according to its `spec`
*/
func (info *Info) Sample(fs *fiatshamir.State) interface{} {
	switch info.Type {
	case Field:
		return fs.RandomField()
	case IntegerVec:
		return fs.RandomManyIntegers(info.Size, info.UpperBound)
	}
	panic("Unreachable")
}

/*
Constructor for Info. For IntegerVec, size[0] contains the
number of integers and size[1] contains the upperBound.
*/
func NewInfo(name Name, type_ Type, round int, size ...int) Info {

	infos := Info{Name: name, Type: type_, Round: round}

	switch type_ {
	case IntegerVec:
		if len(size) != 2 || size[0] < 1 || size[1] < 1 {
			utils.Panic("size was %v", size)
		}
		infos.Size = size[0]
		infos.UpperBound = size[1]
	case Field:
		if len(size) > 0 {
			utils.Panic("size for Field")
		}
	default:
		panic("unreachable")
	}

	return infos
}
