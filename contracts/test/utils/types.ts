export type RawBlockData = {
  rootHash: string;
  timestamp: number;
  rlpEncodedTransactions: string[];
  l2ToL1MsgHashes: string[];
  batchReceptionIndices: number[];
  fromAddresses: string;
};

export type FormattedBlockData = Omit<
  RawBlockData,
  "rlpEncodedTransactions" | "timestamp" | "l2ToL1MsgHashes" | "fromAddresses" | "rootHash"
> & {
  l2BlockTimestamp: number;
  transactions: string[];
  l2ToL1MsgHashes: string[];
  fromAddresses: string;
  blockRootHash: string;
};

export type DebugData = {
  blocks: {
    txHashes: string[];
    hashOfTxHashes: string;
    logHashes: string[];
    hashOfLogHashes: string;
    hashOfPositions: string;
    HashForBlock: string;
  }[];
  hashForAllBlocks: string;
  hashOfRootHashes: string;
  timestampHashes: string;
  finalHash: string;
};
