import { MessageProps } from "../message/Message";
import { MessageSent, OnChainMessageStatus, TransactionResponse } from "../types";
import { IMessageServiceContract, ClaimTransactionOverrides } from "./IMessageServiceContract";

export interface IL1ContractClient extends IMessageServiceContract {
  getMessageStatusUsingMessageHash(messageHash: string): Promise<OnChainMessageStatus>;

  getMessageStatusUsingMerkleTree(params: {
    messageHash: string;
    messageBlockNumber?: number;
  }): Promise<OnChainMessageStatus>;

  estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: { claimViaAddress?: string },
  ): Promise<bigint>;

  estimateClaimWithoutProofGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    opts?: { claimViaAddress?: string },
  ): Promise<bigint>;

  claimWithoutProof(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    opts?: { claimViaAddress?: string; overrides?: ClaimTransactionOverrides },
  ): Promise<TransactionResponse>;
}
