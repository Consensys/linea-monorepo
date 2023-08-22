package json

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/valyala/fastjson"
)

const (
	SIBLINGS   string = "siblings"
	LEAF_INDEX string = "leafIndex"
)

/*
Try parsing a Merkle proof

Example:

	{
	          "leafIndex":2,
	          "siblings":[
	             "0x0000000000000000000000000000000000000000000000000000000000000000",
	             "0x00d0f8b1b21ba489592082a1bb9cd61f9e3394169e1623a3a11a3d60952ca987",
	             "0x173028dc3fc24d89b918ab4952f667ec2f8ea5341ce6c3202b0fefee6cf76041",
	             "0x07cd5828f4e95899b5539065896b855b7684478e6cf2f9e1162ea9a031471846",
	             "0x2e924806f112e278db1694edc2d6b7127053bcddfb24c55c9c47d578c47f0ec8",

						...

				 "0x1c477aa6e33b02996c81cedf8d1e88c9cc6ea40784191e391d23979dcb0208ec",
	             "0x1798747385b7e3752743a21e1dde652c9dee1106d61fa56ccb91cff1e6434884",
	             "0x1254d0207708e41336c9e607dfb109ed26eaa974b75ff9ababcfde3ff70a169a",
	             "0x303b0a04f392124e9264916d8b752241e24f97fa2fa1276d5a7ad197fed3c9c5"
	          ]
	       }
*/
func TryParseMerkleProof(p fastjson.Value, field ...string) (proof smt.Proof, err error) {

	_p := p.Get(field...)
	if _p == nil {
		return smt.Proof{}, fmt.Errorf("not found : %v", field)
	}

	// Safe to dereference
	p = *_p

	// Try getting the leaf index
	leafIndex, err := jsonutil.TryGetInt(p, LEAF_INDEX)
	if err != nil {
		return smt.Proof{}, err
	}
	proof.Path = leafIndex

	// Now collect the proof itself
	array, err := jsonutil.TryGetArray(p, SIBLINGS)
	if err != nil {
		return smt.Proof{}, err
	}

	siblings := make([]Digest, len(array))

	for i := range array {
		// No need to check for nil-ness. It's done ahead of time
		parsed, err := jsonutil.TryGetDigest(array[i])
		if err != nil {
			return smt.Proof{}, err
		}
		siblings[i] = parsed
	}

	proof.Siblings = siblings
	return proof, nil
}
