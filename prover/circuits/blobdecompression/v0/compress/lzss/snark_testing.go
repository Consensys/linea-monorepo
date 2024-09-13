package lzss

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/v0/compress"
	"github.com/consensys/zkevm-monorepo/prover/lib/compressor/blob/v0/compress/lzss"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

type DecompressionTestCircuit struct {
	C                []frontend.Variable
	D                []byte
	Dict             []byte
	CBegin           frontend.Variable
	CLength          frontend.Variable
	CheckCorrectness bool
	Level            lzss.Level
}

func (c *DecompressionTestCircuit) Define(api frontend.API) error {
	dict := utils.ToVariableSlice(lzss.AugmentDict(c.Dict))
	dBack := make([]frontend.Variable, len(c.D)) // TODO Try smaller constants
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
