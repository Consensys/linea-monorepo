package column

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/google/uuid"
)

// PackedStore is a serialization-friendly intermediate struct used for
// serializing a column store.
type PackedStore [][]*storedColumnInfo

// Pack packs the store into a [PackedStore]
func (s *Store[T]) Pack() PackedStore {
	res := make(PackedStore, s.byRounds.Len())
	for i := 0; i < s.byRounds.Len(); i++ {
		res[i] = s.byRounds.MustGet(i)
	}
	return res
}

// TODO @thomas make it a function instead of a method
// func (p PackedStore) Unpack() *Store[T] {
// 	store := NewStore[T]()
// 	for rnd, arr := range p {
// 		for pir, info := range arr {
// 			store.byRounds.AppendToInner(rnd, info)
// 			store.indicesByNames.InsertNew(
// 				info.ID,
// 				columnPosition{
// 					round:      rnd,
// 					posInRound: pir,
// 				},
// 			)
// 		}
// 	}
// 	return store
// }

// PackedNatural is serialization-friendly intermediate structure that is
// use to represent a natural column.
type PackedNatural[T zk.Element] struct {
	Store    *Store[T]    `cbor:"s"`
	Round    int          `cbor:"r"`
	Position int          `cbor:"p"`
	ID       ifaces.ColID `cbor:"i"`
}

// Pack packs a [Natural] into a [PackedNatural]
func (nat Natural[T]) Pack() PackedNatural[T] {
	return PackedNatural[T]{
		Store:    nat.store,
		Round:    nat.position.round,
		Position: nat.position.posInRound,
		ID:       nat.ID,
	}
}

func (unpacked PackedNatural[T]) Unpack() Natural[T] {
	return newNatural[T](
		unpacked.ID,
		columnPosition{round: unpacked.Round, posInRound: unpacked.Position},
		unpacked.Store,
	)
}

// PackedIdentifier returns an identifier that won't conflict with the
// serialization of a [PackedNatural].
func (nat Natural[T]) UUID() uuid.UUID {
	return nat.store.info(nat.ID).uuid
}
