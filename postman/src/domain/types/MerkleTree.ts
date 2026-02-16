export type BlockRange = {
  startingBlock: number;
  endBlock: number;
};

export type FinalizationMessagingInfo = {
  l2MessagingBlocksRange: BlockRange;
  l2MerkleRoots: string[];
  treeDepth: number;
};

export type Proof = {
  proof: string[];
  root: string;
  leafIndex: number;
};
