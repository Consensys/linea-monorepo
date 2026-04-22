//go:build baremetal

package main

func main() {
	input, ok := loadVerifierInput()
	if !ok {
		announceBaremetalInputError()
		haltForever()
	}

	Result = ComputeWords(input.Words)
	if Result == input.Expected {
		announceBaremetalResult(Result)
	} else {
		announceBaremetalMismatch(input.Expected, Result)
	}

	haltForever()
}
