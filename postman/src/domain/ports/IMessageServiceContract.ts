import { MessageProps } from "../message/Message";
import { OnChainMessageStatus, MessageSent, TransactionResponse } from "../types";

export type ClaimTransactionOverrides = {
  nonce?: number;
  gasLimit?: bigint;
  maxFeePerGas?: bigint;
  maxPriorityFeePerGas?: bigint;
};

export interface IMessageServiceContract {
  getMessageStatus(params: { messageHash: string; messageBlockNumber?: number }): Promise<OnChainMessageStatus>;

  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: { claimViaAddress?: string; overrides?: ClaimTransactionOverrides },
  ): Promise<TransactionResponse>;

  retryTransactionWithHigherFee(transactionHash: string, priceBumpPercent?: number): Promise<TransactionResponse>;

  isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean>;

  isRateLimitExceededError(transactionHash: string): Promise<boolean>;

  parseTransactionError(transactionHash: string): Promise<string>;
}
