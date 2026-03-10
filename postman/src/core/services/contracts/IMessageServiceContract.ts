import { MessageProps } from "../../entities/Message";
import { OnChainMessageStatus } from "../../enums";
import { Address, Hash, ErrorDescription, MessageSent, Overrides, TransactionSubmission } from "../../types";

export interface IMessageServiceContract {
  getMessageStatus(params: {
    messageHash: Hash;
    messageBlockNumber?: number;
    overrides?: Overrides;
  }): Promise<OnChainMessageStatus>;
  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: Address; messageBlockNumber?: number },
    opts?: {
      claimViaAddress?: Address;
      overrides?: Overrides;
    },
  ): Promise<TransactionSubmission>;
  retryTransactionWithHigherFee(transactionHash: Hash, priceBumpPercent?: number): Promise<TransactionSubmission>;
  isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean>;
  isRateLimitExceededError(transactionHash: Hash): Promise<boolean>;
  parseTransactionError(transactionHash: Hash): Promise<ErrorDescription | string>;
}
