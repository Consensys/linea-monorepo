package generic

import "github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"

// Definition of the tx-rlp module
var TX_RLP = GenericByteModuleDefinition{
	Data: DataDef{
		HashNum: "txRlp.ABS_TX_NUM",
		Index:   "txRlp.INDEX_LX",
		Limb:    "txRlp.LIMB",
		NBytes:  "txRlp.nBytes",
		LX:      "txRlp.LX",
		LC:      "txRlp.LIMB_CONSTRUCTED",
	},
}

// Definition of the phoney rlp module
var PHONEY_RLP = GenericByteModuleDefinition{
	Data: DataDef{
		HashNum: "phoneyRLP.TX_NUM",
		Index:   "phoneyRLP.INDEX",
		Limb:    "phoneyRLP.LIMB",
		NBytes:  "phoneyRLP.nBYTES",
		// TO_HASH: "phoneyRLP.TO_HASH_BY_PROVER",
	},
}

// Definition of a generic byte module using the column IDs
type GenericByteModuleDefinition struct {
	Data DataDef
	Info InfoDef
}

// DataDef defines the column of a module summarizing informations about the
// data to hash.
type DataDef struct {
	HashNum ifaces.ColID
	Index   ifaces.ColID
	Limb    ifaces.ColID
	NBytes  ifaces.ColID
	LC, LX  ifaces.ColID
	TO_HASH ifaces.ColID
}

// Info module summarizing informations about the hash as a whole
type InfoDef struct {
	HashNum ifaces.ColID
	HashHi  ifaces.ColID
	HashLo  ifaces.ColID
}
