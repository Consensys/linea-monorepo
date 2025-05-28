package column

import (
	"github.com/consensys/linea-monorepo/prover/utils"
)

// PackedStore is a serialization-friendly intermediate struct used for
// serializing a column store.
type PackedStore struct {
	IndicesByName map[string]int `cbor:"n"`
	ByRounds      [][]struct{}   `cbor:"r"`
}

// PackedNatural is serialization-friendly intermediate structure that is
// use to represent a natural column.
type PackedNatural struct {
	Name                string         `cbor:"n"`
	Round               int8           `cbor:"r"`
	Status              int8           `cbor:"s"`
	Log2Size            int8           `cbor:"z"`
	CompiledIOP         int8           `cbor:"i"`
	PosInRound          int8           `cbor:"p"`
	IncludeInProverFS   bool           `cbor:"f"`
	ExcludeFromProverFS bool           `cbor:"e"`
	Pragmas             map[string]any `cbor:"g,omitempty"`
}

// Pack packs a [Natural] into a [PackedNatural]
func (nat Natural) Pack() PackedNatural {

	infos := nat.store.info(nat.ID)

	return PackedNatural{
		Name:                string(nat.ID),
		Round:               int8(nat.Round()),
		Status:              int8(nat.Status()),
		Log2Size:            int8(utils.Log2Ceil(nat.Size())),
		CompiledIOP:         0,
		PosInRound:          int8(nat.position.posInRound),
		IncludeInProverFS:   infos.IncludeInProverFS,
		ExcludeFromProverFS: infos.ExcludeFromProverFS,
		Pragmas:             infos.Pragmas,
	}

}
