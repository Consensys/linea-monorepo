package lzss

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v0/compress"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0/compress/lzss"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type DecompressionTestCircuit struct {
	C                []T
	D                []byte
	Dict             []byte
	CBegin           T
	CLength          T
	CheckCorrectness bool
	Level            lzss.Level
}

func (c *DecompressionTestCircuit) Define(api frontend.API) error {
	dict := utils.ToVariableSlice(lzss.AugmentDict(c.Dict))
	dBack := make([]T, len(c.D)) // TODO Try smaller constants
	if cb, ok := c.CBegin.(int); !ok || cb != 0 {
		c.C = compress.ShiftLeft(api, c.C, c.CBegin)
	}
	dLen, err := Decompress(api, c.C, c.CLength, dBack, dict, c.Level)
	if err != nil {
		return err
	}
	if c.CheckCorrectness {
		api.AssertIsEqual(len(c.D), dLen)
		for i := range c.D {
			api.AssertIsEqual(c.D[i], dBack[i])
		}
	}
	return nil
}
