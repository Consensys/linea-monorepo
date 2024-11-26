import { MessageProps } from "../../entities/Message";
import { OnChainMessageStatus } from "../../enums/MessageEnums";
import { IMessageServiceContract } from "../../services/contracts/IMessageServiceContract";
import { MessageSent } from "../../types/events";
import { FinalizationMessagingInfo, Proof } from "./IMerkleTreeService";

export interface ILineaRollupClient<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse>
  extends IMessageServiceContract<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse> {
  getFinalizationMessagingInfo(transactionHash: string): Promise<FinalizationMessagingInfo>;
  getL2MessageHashesInBlockRange(fromBlock: number, toBlock: number): Promise<string[]>;
  getMessageSiblings(messageHash: string, messageHashes: string[], treeDepth: number): string[];
  getMessageProof(messageHash: string): Promise<Proof>;
  getMessageStatusUsingMessageHash(messageHash: string, overrides: Overrides): Promise<OnChainMessageStatus>;
  getMessageStatusUsingMerkleTree(messageHash: string, overrides: Overrides): Promise<OnChainMessageStatus>;
  estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides?: Overrides,
  ): Promise<bigint>;
  estimateClaimWithoutProofGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<bigint>;
  claimWithoutProof(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<ContractTransactionResponse>;
}
