package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// This are compiler assertions that CoinField implements both the Accessor and
// the Coin interface.
var (
	_ Accessor = &CoinField{}
	_ Coin     = &CoinField{}
)

// CoinField is an implementation of the [Coin] interface representing a
// a random coin value sampled uniformingly over the underlying field.
type CoinField struct {
	round    int
	metadata *metadata
}

// NewCoinField generates a new coin in the API
func (api *API) NewCoinField(round int) *CoinField {

	res := &CoinField{
		round:    round,
		metadata: api.newMetadata(),
	}

	api.coins.addToRound(round, res)
	return res
}

// Round returns the round of interaction where the value of the coin is sampled
// by the verifier.
func (c *CoinField) Round() int {
	return c.round
}

func (c *CoinField) sample(fs *fiatshamir.State) any {
	return fs.RandomField()
}

// GetVal implements the [Accessor] interface
func (c *CoinField) GetVal(run Runtime) field.Element {
	v, ok := run.tryGetCoin(c)
	if !ok {
		utils.Panic("missing value for the coin %v. Explainer: \n%v", c.String(), c.Explain())
	}

	return v.(field.Element)
}

// GetValGnark implements the [Accessor] interface
func (c *CoinField) GetValGnark(api frontend.API, run RuntimeGnark) frontend.Variable {
	v, ok := run.tryGetCoin(c)
	if !ok {
		utils.Panic("missing value for the coin %v. Explainer: \n%v", c.String(), c.Explain())
	}
	return v.(frontend.Variable)
}
