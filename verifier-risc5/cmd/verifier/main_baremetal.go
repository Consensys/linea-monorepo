//go:build baremetal

package main

func main() {
	input, ok := loadVerifierInput()
	if !ok {
		announceBaremetalInputError()
		haltForever()
	}

	var okCompute bool
	Result, okCompute = ComputeWordsChecked(input.Words)
	if !okCompute {
		announceBaremetalInputError()
		haltForever()
	}

	if Result == input.Expected {
		announceBaremetalResult(Result)
	} else {
		announceBaremetalMismatch(input.Expected, Result)
	}

	haltForever()
}
