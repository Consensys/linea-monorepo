package keccak

// Settings is a collection of parameters that we use to instantiate the
// keccak module.
type Settings struct {
	Enabled       bool // will be zero if the option is not called
	MaxNumKeccakf int
}
