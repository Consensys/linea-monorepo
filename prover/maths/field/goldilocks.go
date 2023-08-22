//go:build goldilocks

package field

import "github.com/consensys/gnark-crypto/field/goldilocks"

// Flag indicating whether we are using goldilocks
const USING_GOLDILOCKS = true

// Type alias to make it easy to switch
type Element = goldilocks.Element

// Function alias for the butterfly
var Butterfly = goldilocks.Butterfly

var One = goldilocks.One

var NewElement = goldilocks.NewElement

const Bits = goldilocks.Bits
const Bytes = goldilocks.Bytes

var BatchInvert = goldilocks.BatchInvert

const RootOfUnity string = "2741030659394132017"
const RootOrUnityOrder uint64 = 32
const MultiplicativeGen uint64 = 11
