package json

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/accumulator"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/valyala/fastjson"
)

// Sugar alias type
type (
	LeafOpening = accumulator.LeafOpening
	Digest      = hashtypes.Digest
)

// JSON field names
const (
	HKEY      string = "hkey"
	HVAL      string = "hval"
	PREV_LEAF string = "prevLeaf"
	NEXT_LEAF string = "nextLeaf"
)

// Attempt to parse a leaf opening from a fastjson parse
func TryParseLeafOpening(p fastjson.Value, field ...string) (LeafOpening, error) {

	// Check the json path
	_p := p.Get(field...)
	if _p == nil {
		return LeafOpening{}, fmt.Errorf("not found : %v", field)
	}

	// Safe to dereference
	p = *_p

	lo := LeafOpening{}
	var err error

	lo.HKey, err = jsonutil.TryGetDigest(p, HKEY)
	if err != nil {
		return LeafOpening{}, err
	}

	lo.HVal, err = jsonutil.TryGetDigest(p, HVAL)
	if err != nil {
		return LeafOpening{}, err
	}

	lo.Prev, err = jsonutil.TryGetInt64(p, PREV_LEAF)
	if err != nil {
		return LeafOpening{}, err
	}

	lo.Next, err = jsonutil.TryGetInt64(p, NEXT_LEAF)
	if err != nil {
		return LeafOpening{}, err
	}

	return lo, nil
}
