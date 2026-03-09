import { MessageProps } from "../../entities/Message";
import { OnChainMessageStatus } from "../../enums";
import { ErrorDescription, MessageSent, Overrides, TransactionSubmission } from "../../types";

export interface IMessageServiceContract {
  getMessageStatus(params: {
    messageHash: string;
    messageBlockNumber?: number;
    overrides?: Overrides;
  }): Promise<OnChainMessageStatus>;
  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: {
      claimViaAddress?: string;
      overrides?: Overrides;
    },
  ): Promise<TransactionSubmission>;
  retryTransactionWithHigherFee(transactionHash: string, priceBumpPercent?: number): Promise<TransactionSubmission>;
  isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean>;
  isRateLimitExceededError(transactionHash: string): Promise<boolean>;
  parseTransactionError(transactionHash: string): Promise<ErrorDescription | string>;
}
