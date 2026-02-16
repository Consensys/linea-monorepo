import { FinalizationMessagingInfo, Proof } from "../types";

export interface IMerkleTreeService {
  getMessageProof(messageHash: string): Promise<Proof>;
  getFinalizationMessagingInfo(transactionHash: string): Promise<FinalizationMessagingInfo>;
  getL2MessageHashesInBlockRange(fromBlock: number, toBlock: number): Promise<string[]>;
  getMessageSiblings(messageHash: string, messageHashes: string[], treeDepth: number): string[];
}
