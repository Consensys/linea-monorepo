package column

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

// PackedStoreFlat is the serialization-friendly flattened version of PackedStore.
// Instead of storing [][]*storedColumnInfo (nested slices with embedded pointers),
// we flatten it into:
// - RoundLengths: stores the length of each round's inner slice
// - Data: all storedColumnInfo values linearized in round order

// PackedNatural is serialization-friendly intermediate structure that is
// used to represent a natural column.
type PackedNatural struct {
	Store    *Store       `cbor:"s"`
	Round    int          `cbor:"r"`
	Position int          `cbor:"p"`
	ID       ifaces.ColID `cbor:"i"`
}

// Pack packs a [Natural] into a [PackedNatural]
func (nat Natural) Pack() PackedNatural {
	return PackedNatural{
		Store:    nat.store,
		Round:    nat.position.round,
		Position: nat.position.posInRound,
		ID:       nat.ID,
	}
}

func (unpacked PackedNatural) Unpack() Natural {
	return newNatural(
		unpacked.ID,
		columnPosition{round: unpacked.Round, posInRound: unpacked.Position},
		unpacked.Store,
	)
}

type PackedStoreFlat struct {
	RoundLengths []int              `cbor:"rl"`
	Data         []storedColumnInfo `cbor:"d"`
}

func (s *Store) Pack() PackedStoreFlat {
	numRounds := s.byRounds.Len()
	roundLengths := make([]int, numRounds)
	var data []storedColumnInfo

	for rnd := 0; rnd < numRounds; rnd++ {
		roundData := s.byRounds.MustGet(rnd) // []*storedColumnInfo
		roundLengths[rnd] = len(roundData)
		for _, ptr := range roundData {
			data = append(data, *ptr) // dereference to storedColumnInfo
		}
	}

	return PackedStoreFlat{
		RoundLengths: roundLengths,
		Data:         data,
	}
}

func (p PackedStoreFlat) Unpack() *Store {
	store := NewStore()

	dataIdx := 0
	for rnd, roundLen := range p.RoundLengths {
		for posInRound := 0; posInRound < roundLen; posInRound++ {
			info := p.Data[dataIdx]
			dataIdx++

			// reallocate as pointer because byRounds stores []*storedColumnInfo
			copied := info
			store.byRounds.AppendToInner(rnd, &copied)
			store.indicesByNames.InsertNew(
				info.ID,
				columnPosition{
					round:      rnd,
					posInRound: posInRound,
				},
			)
		}
	}
	return store
}
