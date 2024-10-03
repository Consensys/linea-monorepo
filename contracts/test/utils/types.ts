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

export type SubmissionData = {
  finalStateRootHash: string;
  firstBlockInData: bigint;
  finalBlockInData: bigint;
  snarkHash: string;
};

export type BlobSubmissionData = {
  submissionData: SubmissionData;
  dataEvaluationClaim: string;
  kzgCommitment: string;
  kzgProof: string;
};

export type ParentSubmissionData = {
  finalStateRootHash: string;
  firstBlockInData: bigint;
  finalBlockInData: bigint;
  shnarf: string;
};

export type ParentAndExpectedShnarf = {
  parentShnarf: string;
  expectedShnarf: string;
};

export type ShnarfData = {
  parentShnarf: string;
  snarkHash: string;
  finalStateRootHash: string;
  dataEvaluationPoint: string;
  dataEvaluationClaim: string;
};

export type CalldataSubmissionData = SubmissionData & {
  compressedData: string;
};

export type SubmissionAndCompressedData = {
  submissionData: SubmissionData;
  compressedData: string;
};

export type FinalizationData = {
  aggregatedProof: string;
  finalBlockInData: bigint;
  shnarfData: ShnarfData;
  parentStateRootHash: string;
  lastFinalizedTimestamp: bigint;
  finalTimestamp: bigint;
  l1RollingHash: string;
  l1RollingHashMessageNumber: bigint;
  l2MerkleRoots: string[];
  l2MerkleTreesDepth: bigint;
  l2MessagingBlocksOffsets: string;
  lastFinalizedL1RollingHash: string;
  lastFinalizedL1RollingHashMessageNumber: bigint;
};
