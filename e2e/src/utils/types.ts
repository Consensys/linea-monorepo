import { BigNumber, BytesLike } from "ethers";

export type FinalizationData = {
  aggregatedProof: string;
  aggregatedProverVersion: string;
  aggregatedVerifierIndex: number;
  aggregatedProofPublicInput: string;
  dataHashes: string[];
  dataParentHash: string;
  parentStateRootHash: string;
  parentAggregationLastBlockTimestamp: number;
  finalTimestamp: number;
  finalBlockNumber: number;
  l1RollingHash: string;
  l1RollingHashMessageNumber: number;
  l2MerkleRoots: string[];
  l2MerkleTreesDepth: number;
  l2MessagingBlocksOffsets: string;
};

export type MessageEvent = {
  from: string;
  to: string;
  fee: BigNumber;
  value: BigNumber;
  messageNumber: BigNumber;
  calldata: string;
  messageHash: string;
  blockNumber: number;
};

export type SendMessageArgs = {
  to: string;
  fee: BigNumber;
  calldata: BytesLike;
};
