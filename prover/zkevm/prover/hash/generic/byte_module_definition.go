package generic

import "github.com/consensys/linea-monorepo/prover/protocol/ifaces"

// Definition of the tx-rlp module
var RLP_TXN = GenericByteModuleDefinition{
	Data: DataDef{
		HashNum: "txRlp.ABS_TX_NUM",
		Index:   "txRlp.INDEX_LX",
		Limb:    "txRlp.LIMB",
		NBytes:  "txRlp.nBytes",
		TO_HASH: "txRlp.ToHash",
	},
}

// Definition of the shakira module
var SHAKIRA = GenericByteModuleDefinition{
	Data: DataDef{
		HashNum: "shakira.TX_NUM",
		Index:   "shakira.INDEX",
		Limb:    "shakira.LIMB",
		NBytes:  "shakira.nBYTES",
		TO_HASH: "shakira.TO_HASH_BY_PROVER",
	},
	Info: InfoDef{
		HashNum:  "shakira.TX_NUM_Info",
		HashLo:   "shakira_HashLo",
		HashHi:   "shakira.HashHi",
		IsHashLo: "shakira.IsHashLo",
		IsHashHi: "shakira.IsHashHi",
	},
}

// Definition of the shakira module
var RLP_ADD = GenericByteModuleDefinition{
	Data: DataDef{
		HashNum: "rlp_addr.TX_NUM",
		Index:   "rlp_addr.INDEX",
		Limb:    "rlp_addr.LIMB",
		NBytes:  "rlp_addr.nBYTES",
		TO_HASH: "rlp_addr.TO_HASH_BY_PROVER",
	},
	Info: InfoDef{
		HashNum:  "rlp_addr.TX_NUM_Info",
		HashLo:   "rlp_addr.HashLo",
		HashHi:   "rlp_addr.HashHi",
		IsHashLo: "rlp_addr.IsHashLo",
		IsHashHi: "rlp_addr.IsHashHi",
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

// DataDef defines the column of a module summarizing information about the
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
	HashNum  ifaces.ColID
	HashLo   ifaces.ColID
	HashHi   ifaces.ColID
	IsHashLo ifaces.ColID
	IsHashHi ifaces.ColID
}
