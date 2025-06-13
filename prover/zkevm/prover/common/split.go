package common

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// LimbBytes is the size of one limb in bytes
const LimbBytes = 2

// SplitElement splits the input field element into subarrays of the provided size.
func SplitElement(element field.Element) []field.Element {
	input := element.Bytes()

	var result []field.Element
	for i := 0; i < len(input); i += LimbBytes {
		end := i + LimbBytes
		if end > len(input) {
			end = len(input)
		}

		var limb field.Element
		limb.SetBytes(input[i:end])
		result = append(result, limb)
	}
	return result
}

func CombineElements(elements []field.Element) field.Element {
	var bytes []byte
	for _, element := range elements {
		elementBytes := element.Bytes()
		bytes = append(bytes, elementBytes[len(elementBytes)-LimbBytes:]...)
	}

	var res field.Element
	res.SetBytes(bytes)

	return res
}
