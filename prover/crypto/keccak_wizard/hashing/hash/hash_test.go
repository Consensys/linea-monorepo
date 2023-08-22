package hash

import (
	"bytes"
	"hash"
	"testing"
)

func TestKeccak(t *testing.T) {
	tests := []struct {
		fn   func() hash.Hash
		data []byte
		want string
	}{
		{
			NewLegacyKeccak256,
			[]byte{0x00},
			"bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a",
		},
		{
			NewLegacyKeccak256,
			[]byte{0x00, 0x00},
			"54a8c0ab653c15bfb48b47fd011ba2b9617af01cb45cab344acd57c924d56798",
		},
		{
			NewLegacyKeccak256,
			[]byte("abc"),
			"4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45",
		},
		{
			NewLegacyKeccak256,
			[]byte("Test Case: we test a message of length 136 bytes, namely, equal to the Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate!"),
			"65884b747b16d83ef56bbc088829fa66dad5e3ee2dd3769fa9bcc92e8ce0cbad",
		},
		{
			NewLegacyKeccak256,
			[]byte("Test Case: we test a message with the length larger than 136 bytes, namely, larger than the Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate Rate!"),
			"054d7fc3ae17bd0cddcba0027b7c335d8a8412f16309acc9b435b8f1c1c10a0e",
		},
	}

	h := NewLegacyKeccak256()
	//the number of permutations
	//trcaker to define a type for data

	for _, u := range tests {

		//h := u.fn()
		h.Reset()
		h.Write(u.data)
		got := h.Sum(nil)
		want := DecodeHex(u.want)

		if !bytes.Equal(got, want) {
			t.Errorf("unexpected hash for size %d: got '%x' want '%s'", h.Size()*8, got, u.want)
		}

	}

}
